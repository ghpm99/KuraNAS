import { FileType } from '@/utils';
import { useInfiniteQuery, useMutation, useQuery } from '@tanstack/react-query';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import {
	copyFilePath,
	createFolderAtPath,
	deleteFilePath,
	getFileByPath,
	getFilesTree,
	getRecentAccessByFileId,
	moveFilePath,
	renameFilePath,
	rescanFiles as requestFilesRescan,
	toggleStarredFile,
	uploadFilesToPath,
} from '@/service/files';
import {
	FileContextProvider,
	FileContextType,
	FileData,
	FileListCategoryType,
	PaginationResponse,
} from './fileContext';
import {
	addChildrenToTree,
	buildFilesUrl,
	extractFilePath,
	findItemInTree,
	findTrailByIdInTree,
} from './fileProviderUtils';

const pageSize = 200;

const FileProvider = ({ children }: { children: React.ReactNode }) => {
	const location = useLocation();
	const navigate = useNavigate();

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

	// Snapshot derived from resolvedItem — no state/effect needed
	const selectedItemSnapshot = currentFilePath ? (resolvedItem ?? null) : null;

	const queryParams = useMemo(
		() => ({
			page_size: pageSize,
			file_parent: selectedItemId ?? undefined,
		}),
		[selectedItemId],
	);

	const { status, data, refetch } = useInfiniteQuery({
		queryKey: ['files', queryParams, fileListFilter],
		queryFn: ({ pageParam = 1 }): Promise<PaginationResponse> =>
			getFilesTree({
				page: pageParam,
				pageSize,
				fileParent: selectedItemId ?? undefined,
				category: fileListFilter,
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

	const rescanFiles = useCallback(async () => {
		await requestFilesRescan();
		await refetch();
	}, [refetch]);

	const uploadFiles = useCallback(
		async (files: FileList, targetPath?: string) => {
			await uploadFilesToPath(files, targetPath);
			await refetch();
		},
		[refetch],
	);

	const createFolder = useCallback(
		async (name: string, parentPath?: string) => {
			await createFolderAtPath(name, parentPath);
			await refetch();
		},
		[refetch],
	);

	const movePath = useCallback(
		async (sourcePath: string, destinationPath: string) => {
			await moveFilePath(sourcePath, destinationPath);
			await refetch();
		},
		[refetch],
	);

	const copyPath = useCallback(
		async (sourcePath: string, destinationPath: string) => {
			await copyFilePath(sourcePath, destinationPath);
			await refetch();
		},
		[refetch],
	);

	const renamePath = useCallback(
		async (sourcePath: string, newName: string) => {
			await renameFilePath(sourcePath, newName);
			await refetch();
		},
		[refetch],
	);

	const deletePath = useCallback(
		async (path: string) => {
			await deleteFilePath(path);
			await refetch();
		},
		[refetch],
	);

	// Update file tree when data arrives (deferred to avoid cascading renders)
	useEffect(() => {
		if (!data) return;
		const nextItems = data?.pages[0]?.items ?? [];
		let cancelled = false;
		if (selectedItemId) {
			queueMicrotask(() => {
				if (cancelled) return;
				setFileTree((currentTree) => addChildrenToTree(currentTree, selectedItemId, nextItems));
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
	}, [data, selectedItemId]);

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
			const nextItems = data.pages[0]?.items ?? [];
			return { ...selectedItemSnapshot, file_children: nextItems };
		}

		return selectedItemSnapshot;
	}, [selectedItemId, fileTree, selectedItemSnapshot, data]);

	// Navigate via URL (push for browser history)
	const handleSelectItem = useCallback(
		(item: FileData | null) => {
			if (!item) {
				navigate(buildFilesUrl(''));
				return;
			}
			navigate(buildFilesUrl(item.path));
		},
		[navigate],
	);

	const handleStarredItem = useCallback(
		(itemId: number) => {
			updateStarredFile(itemId);
		},
		[updateStarredFile],
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
			handleStarredItem,
			uploadFiles,
			createFolder,
			movePath,
			copyPath,
			renamePath,
			deletePath,
			rescanFiles,
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
			handleStarredItem,
			uploadFiles,
			createFolder,
			movePath,
			copyPath,
			renamePath,
			deletePath,
			rescanFiles,
		],
	);
	return <FileContextProvider value={contextValue}>{children}</FileContextProvider>;
};

export default FileProvider;
