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
    setFileStarred,
    toggleStarredFile,
} from '@/service/files';
import { promoteFileToHot as promoteFileToHotService } from '@/service/tiering';
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
    extractFilePath,
    findItemInTree,
    findTrailByIdInTree,
} from './fileProviderUtils';
import { buildBrowserUrl, getFileBrowserRootPath } from '@/app/routes';
import useExpandTreeAlongAncestors from './useExpandTreeAlongAncestors';
import { extractBackendErrorMessage } from '../../fileActions/bulkOutcome';
import { allFileQueryKeys, searchQueryKey } from '@/shared/queryKeys/fileQueryKeys';
import { listingStaleTimeMs } from '@/components/providers/queryFreshness';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useSnackbar } from 'notistack';
import {
    applyStarredToFile,
    applyStarredToListing,
    type FileListingCache,
} from './fileStarredCache';

const pageSize = 200;

const listingQueryKey = ['files'] as const;
const openedItemQueryKey = ['files-path'] as const;

type StarredVariables = { fileId: number; isStarred: boolean };

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
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();

    const browserRootPath = getFileBrowserRootPath(location.pathname);
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

    const { status, error, data, refetch, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
        queryKey: ['files', queryParams, fileListFilter, filesSort],
        queryFn: ({ pageParam = 1 }): Promise<PaginationResponse> =>
            isGlobalListing(fileListFilter, selectedItemId)
                ? fetchGlobalListing(fileListFilter, pageParam)
                : getFilesTree({
                      page: pageParam,
                      pageSize,
                      fileParent: selectedItemId ?? undefined,
                      category: selectedItemId ? 'all' : fileListFilter,
                      sort: filesSort,
                  }),
        initialPageParam: 1,
        getNextPageParam: (lastPage) => {
            if (lastPage.pagination.hasNext) {
                return lastPage.pagination.page + 1;
            }
            return undefined;
        },
        staleTime: listingStaleTimeMs,
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

    const applyStarredToCaches = useCallback(
        (fileId: number, isStarred: boolean) => {
            queryClient
                .getQueriesData<FileListingCache>({ queryKey: listingQueryKey })
                .forEach(([queryKey, cache]) => {
                    queryClient.setQueryData(
                        queryKey,
                        applyStarredToListing(cache, queryKey, fileId, isStarred)
                    );
                });
            queryClient
                .getQueriesData<FileData | null>({ queryKey: openedItemQueryKey })
                .forEach(([queryKey, cachedFile]) => {
                    queryClient.setQueryData(
                        queryKey,
                        applyStarredToFile(cachedFile, fileId, isStarred)
                    );
                });
        },
        [queryClient]
    );

    const { mutate: updateStarredFile } = useMutation({
        mutationFn: ({ fileId, isStarred }: StarredVariables) => setFileStarred(fileId, isStarred),
        onMutate: async ({ fileId, isStarred }) => {
            await queryClient.cancelQueries({ queryKey: listingQueryKey });
            await queryClient.cancelQueries({ queryKey: openedItemQueryKey });
            const listingSnapshots = queryClient.getQueriesData({ queryKey: listingQueryKey });
            const openedItemSnapshots = queryClient.getQueriesData({
                queryKey: openedItemQueryKey,
            });
            applyStarredToCaches(fileId, isStarred);
            return { snapshots: [...listingSnapshots, ...openedItemSnapshots] };
        },
        onSuccess: (starredState) => {
            applyStarredToCaches(starredState.id, starredState.starred);
        },
        onError: (error, _variables, context) => {
            context?.snapshots.forEach(([queryKey, cache]) => {
                queryClient.setQueryData(queryKey, cache);
            });
            enqueueSnackbar(extractBackendErrorMessage(error) ?? t('ERROR_FAVORITE_FAILED'), {
                variant: 'error',
            });
        },
        onSettled: () => {
            queryClient.invalidateQueries({ queryKey: listingQueryKey });
            queryClient.invalidateQueries({ queryKey: [searchQueryKey] });
        },
    });

    const invalidateFileQueries = useCallback(async () => {
        await Promise.all(
            allFileQueryKeys.map((queryKey) =>
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

    const promoteFileToHot = useCallback(
        async (id: number) => {
            await promoteFileToHotService(id);
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
                navigate(buildBrowserUrl(browserRootPath, movedPath));
            }
            await invalidateFileQueries();
        },
        [openedItemId, browserRootPath, discardOpenedItemPathQuery, navigate, invalidateFileQueries]
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
                    buildBrowserUrl(
                        browserRootPath,
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
            browserRootPath,
            discardOpenedItemPathQuery,
            navigate,
            invalidateFileQueries,
        ]
    );

    const deleteFile = useCallback(
        async (id: number, permanent = false) => {
            await deleteFileService(id, permanent);
            if (id === openedItemId) {
                discardOpenedItemPathQuery();
                navigate(
                    buildBrowserUrl(
                        browserRootPath,
                        openedItemParentPath === '/' ? '' : (openedItemParentPath ?? '')
                    )
                );
            }
            await invalidateFileQueries();
        },
        [
            openedItemId,
            openedItemParentPath,
            browserRootPath,
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
                navigate(buildBrowserUrl(browserRootPath, ''));
                return;
            }
            navigate(buildBrowserUrl(browserRootPath, item.path));
        },
        [navigate, browserRootPath]
    );

    const handleStarredItem = useCallback(
        (itemId: number) => {
            const currentFile =
                findItemInTree(fileTree, itemId) ??
                (selectedItemSnapshot?.id === itemId ? selectedItemSnapshot : null);
            updateStarredFile({ fileId: itemId, isStarred: !currentFile?.starred });
        },
        [updateStarredFile, fileTree, selectedItemSnapshot]
    );

    const contextValue: FileContextType = useMemo(
        () => ({
            files: fileTree || [],
            status: status,
            listingErrorMessage: extractBackendErrorMessage(error),
            retryListing: () => {
                refetch();
            },
            browserRootPath,
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
            promoteFileToHot,
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
            error,
            refetch,
            browserRootPath,
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
            promoteFileToHot,
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
