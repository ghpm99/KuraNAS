import { FileType } from '@/utils';
import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import {
    copyFile as copyFileService,
    createFolder as createFolderService,
    deleteFile as deleteFileService,
    getFileByPath,
    getFilesTree,
    getRecentAccessByFileId,
    getRecentlyAccessedFiles,
    getStarredFiles,
    moveFile as moveFileService,
    renameFile as renameFileService,
    rescanFiles as requestFilesRescan,
    toggleStarredFile,
} from '@/service/files';
import {
    FileContextProvider,
    FileContextType,
    FileData,
    FileListCategoryType,
    FilesSort,
    PaginationResponse,
} from './fileContext';
import FileSelectionProvider from '../../selection/fileSelectionProvider';
import { loadFilesSort, saveFilesSort } from './filesSortPreference';
import {
    addChildrenToTree,
    buildFilesUrl,
    extractFilePath,
    fileQueryKeys,
    findItemInTree,
    findTrailByIdInTree,
} from './fileProviderUtils';
import useExpandTreeAlongAncestors from './useExpandTreeAlongAncestors';

const pageSize = 200;
const ancestorsQueryKey = 'files-ancestors';

const joinPath = (parentPath: string | undefined, name: string | undefined) =>
    `${parentPath === '/' ? '' : (parentPath ?? '')}/${name ?? ''}`;

const isGlobalListing = (filter: FileListCategoryType, parentId: number | null) =>
    parentId === null && filter !== 'all';

const fetchGlobalListing = (filter: FileListCategoryType, page: number): Promise<PaginationResponse> => {
    const params = { page, pageSize };
    return filter === 'starred' ? getStarredFiles(params) : getRecentlyAccessedFiles(params);
};

const FileProvider = ({ children }: { children: React.ReactNode }) => {
    const location = useLocation();
    const navigate = useNavigate();
    const queryClient = useQueryClient();

    // URL → path extraction
    const currentFilePath = extractFilePath(location.pathname);

    // Resolve URL path → FileData via API
    const { data: resolvedItem } = useQuery({
        queryKey: ['files-path', currentFilePath],
        queryFn: () => getFileByPath(currentFilePath),
        enabled: currentFilePath.length > 0,
        staleTime: 30_000,
    });

    // selectedItemId is derived from URL resolution
    const selectedItemId = currentFilePath ? (resolvedItem?.id ?? null) : null;

    const [fileTree, setFileTree] = useState<FileData[]>([]);
    const [fileListFilter, setFileListFilter] = useState<FileListCategoryType>('all');
    const [filesSort, setFilesSortState] = useState<FilesSort>(loadFilesSort);

    const setFilesSort = useCallback((nextSort: FilesSort) => {
        saveFilesSort(nextSort);
        setFilesSortState(nextSort);
    }, []);

    useExpandTreeAlongAncestors({ selectedItemId, filesSort, setFileTree });

    // Snapshot derived from resolvedItem — no state/effect needed
    const selectedItemSnapshot = currentFilePath ? (resolvedItem ?? null) : null;

    const queryParams = useMemo(
        () => ({
            page_size: pageSize,
            file_parent: selectedItemId ?? undefined,
        }),
        [selectedItemId]
    );

    const { status, data, refetch, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
        queryKey: ['files', queryParams, fileListFilter, filesSort],
        queryFn: ({ pageParam = 1 }): Promise<PaginationResponse> =>
            isGlobalListing(fileListFilter, selectedItemId)
                ? fetchGlobalListing(fileListFilter, pageParam)
                : getFilesTree({
                      page: pageParam,
                      pageSize,
                      fileParent: selectedItemId ?? undefined,
                      category: fileListFilter,
                      sort: filesSort,
                  }),
        initialPageParam: 1,
        getNextPageParam: (lastPage) => {
            if (lastPage.pagination.hasNext) {
                return lastPage.pagination.page + 1;
            }
            return undefined;
        },
        staleTime: 0,
    });

    const loadedItems = useMemo(
        () => data?.pages.flatMap((page) => page?.items ?? []) ?? [],
        [data]
    );

    const { data: fileAccessData, isLoading: isLoadingAccessData } = useQuery({
        queryKey: ['filesRecent', 'tree', selectedItemId],
        queryFn: async () => {
            if (!selectedItemId) return [];
            const fromTree = findItemInTree(fileTree, selectedItemId);
            const item = fromTree ?? selectedItemSnapshot;
            if (item?.type !== FileType.File) return [];

            return getRecentAccessByFileId(selectedItemId);
        },
        staleTime: 0,
    });

    const { mutate: updateStarredFile } = useMutation({
        mutationFn: (itemId: number) => toggleStarredFile(itemId),
        onSuccess: () => {
            refetch();
        },
    });

    const invalidateFileQueries = useCallback(async () => {
        await Promise.all(
            [...fileQueryKeys, ancestorsQueryKey].map((queryKey) =>
                queryClient.invalidateQueries({ queryKey: [queryKey] })
            )
        );
    }, [queryClient]);

    const toggleStarred = useCallback(
        async (itemId: number) => {
            await toggleStarredFile(itemId);
            await invalidateFileQueries();
        },
        [invalidateFileQueries]
    );

    const rescanFiles = useCallback(async () => {
        await requestFilesRescan();
        await refetch();
    }, [refetch]);

    const createFolder = useCallback(
        async (name: string, parentId?: number) => {
            await createFolderService(name, parentId);
            await refetch();
        },
        [refetch]
    );

    const openedItemId = currentFilePath ? resolvedItem?.id : undefined;
    const openedItemParentPath = resolvedItem?.parent_path;
    const openedItemName = resolvedItem?.name;

    const discardOpenedItemPathQuery = useCallback(() => {
        queryClient.removeQueries({ queryKey: ['files-path', currentFilePath] });
    }, [queryClient, currentFilePath]);

    const moveFile = useCallback(
        async (sourceId: number, destinationFolderId?: number, destinationPath?: string) => {
            const movedPath = await moveFileService(sourceId, destinationFolderId, destinationPath);
            if (sourceId === openedItemId) {
                discardOpenedItemPathQuery();
                navigate(buildFilesUrl(movedPath));
            }
            await invalidateFileQueries();
        },
        [openedItemId, discardOpenedItemPathQuery, navigate, invalidateFileQueries]
    );

    const copyFile = useCallback(
        async (sourceId: number, destinationFolderId?: number, destinationPath?: string, newName?: string) => {
            await copyFileService(sourceId, destinationFolderId, destinationPath, newName);
            await refetch();
        },
        [refetch]
    );

    const renameFile = useCallback(
        async (id: number, newName: string) => {
            const renamedPath = await renameFileService(id, newName);
            if (id === openedItemId) {
                discardOpenedItemPathQuery();
                navigate(
                    buildFilesUrl(
                        renamedPath || joinPath(openedItemParentPath, newName || openedItemName)
                    )
                );
            }
            await invalidateFileQueries();
        },
        [
            openedItemId,
            openedItemParentPath,
            openedItemName,
            discardOpenedItemPathQuery,
            navigate,
            invalidateFileQueries,
        ]
    );

    const deleteFile = useCallback(
        async (id: number) => {
            await deleteFileService(id);
            if (id === openedItemId) {
                discardOpenedItemPathQuery();
                navigate(buildFilesUrl(openedItemParentPath === '/' ? '' : (openedItemParentPath ?? '')));
            }
            await invalidateFileQueries();
        },
        [
            openedItemId,
            openedItemParentPath,
            discardOpenedItemPathQuery,
            navigate,
            invalidateFileQueries,
        ]
    );

    // Update file tree when data arrives (deferred to avoid cascading renders)
    useEffect(() => {
        if (!data) return;
        const nextItems = loadedItems;
        let cancelled = false;
        if (selectedItemId) {
            queueMicrotask(() => {
                if (cancelled) return;
                setFileTree((currentTree) =>
                    addChildrenToTree(currentTree, selectedItemId, nextItems)
                );
            });
            return () => {
                cancelled = true;
            };
        }
        queueMicrotask(() => {
            if (!cancelled) {
                setFileTree(nextItems);
            }
        });
        return () => {
            cancelled = true;
        };
    }, [data, loadedItems, selectedItemId]);

    // Compute expanded items from the selected item's trail in the tree (derived, not state)
    const expandedItems = useMemo(() => {
        if (!selectedItemId) return [];
        const trail = findTrailByIdInTree(fileTree, selectedItemId);
        if (trail && trail.length > 0) {
            return trail.map((item) => item.id);
        }
        return [];
    }, [selectedItemId, fileTree]);

    // Build effective selected item: tree lookup → snapshot with children
    const effectiveSelectedItem = useMemo(() => {
        if (!selectedItemId) return null;

        const fromTree = findItemInTree(fileTree, selectedItemId);
        if (fromTree) return fromTree;

        if (selectedItemSnapshot && selectedItemSnapshot.type === FileType.Directory && data) {
            return { ...selectedItemSnapshot, file_children: loadedItems };
        }

        return selectedItemSnapshot;
    }, [selectedItemId, fileTree, selectedItemSnapshot, data, loadedItems]);

    // Navigate via URL (push for browser history)
    const handleSelectItem = useCallback(
        (item: FileData | null) => {
            if (!item) {
                navigate(buildFilesUrl(''));
                return;
            }
            navigate(buildFilesUrl(item.path));
        },
        [navigate]
    );

    const handleStarredItem = useCallback(
        (itemId: number) => {
            updateStarredFile(itemId);
        },
        [updateStarredFile]
    );

    const contextValue: FileContextType = useMemo(
        () => ({
            files: fileTree || [],
            status: status,
            selectedItem: effectiveSelectedItem,
            handleSelectItem,
            expandedItems,
            recentAccessFiles: fileAccessData || [],
            isLoadingAccessData: isLoadingAccessData,
            fileListFilter,
            setFileListFilter,
            filesSort,
            setFilesSort,
            handleStarredItem,
            toggleStarred,
            createFolder,
            moveFile,
            copyFile,
            renameFile,
            deleteFile,
            rescanFiles,
            fetchNextPage: () => {
                fetchNextPage();
            },
            hasNextPage: Boolean(hasNextPage),
            isFetchingNextPage: Boolean(isFetchingNextPage),
        }),
        [
            fileTree,
            status,
            effectiveSelectedItem,
            handleSelectItem,
            expandedItems,
            fileAccessData,
            isLoadingAccessData,
            fileListFilter,
            filesSort,
            setFilesSort,
            handleStarredItem,
            toggleStarred,
            createFolder,
            moveFile,
            copyFile,
            renameFile,
            deleteFile,
            rescanFiles,
            fetchNextPage,
            hasNextPage,
            isFetchingNextPage,
        ]
    );
    const selectionScopeKey = `${selectedItemId ?? 'root'}:${fileListFilter}`;

    return (
        <FileContextProvider value={contextValue}>
            <FileSelectionProvider scopeKey={selectionScopeKey}>{children}</FileSelectionProvider>
        </FileContextProvider>
    );
};

export default FileProvider;
