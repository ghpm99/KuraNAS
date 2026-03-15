import { FileType } from '@/utils';
import { useInfiniteQuery, useMutation, useQuery } from '@tanstack/react-query';
import { useCallback, useEffect, useMemo, useState } from 'react';
import {
	copyFilePath,
	createFolderAtPath,
	deleteFilePath,
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

const pageSize = 200;

const findItemInTree = (data: FileData[], itemId: number | null): FileData | null => {
	if (!itemId) return null;
	for (const item of data) {
		if (item.id === itemId) {
			return item;
		}
		if (item?.file_children && item?.file_children?.length > 0) {
			const itemChildren = findItemInTree(item?.file_children, itemId);
			if (itemChildren) {
				return itemChildren;
			}
		}
	}

	return null;
};

const addChildrenToTree = (tree: FileData[], parentId: number, children?: FileData[]): FileData[] => {
	return tree.map((node) => {
		if (node.id === parentId) {
			return { ...node, file_children: children };
		}
		if (node.file_children) {
			return { ...node, file_children: addChildrenToTree(node.file_children, parentId, children) };
		}
		return node;
	});
};

const FileProvider = ({ children }: { children: React.ReactNode }) => {
	const [selectedItemId, setSelectedItemId] = useState<number | null>(null);
	const [fileTree, setFileTree] = useState<FileData[]>([]);
	const [expandedItems, setExpandedItems] = useState<number[]>([]);
	const [fileListFilter, setFileListFilter] = useState<FileListCategoryType>('all');
	const [selectedItemSnapshot, setSelectedItemSnapshot] = useState<FileData | null>(null);

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
				if (selectedItem?.type !== FileType.File || selectedItemId == null) return [];

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

	useEffect(() => {
		if (!data) return;
		const nextItems = data?.pages[0]?.items ?? [];
		if (selectedItemId) {
			// eslint-disable-next-line react-hooks/set-state-in-effect
			setFileTree((currentTree) => addChildrenToTree(currentTree, selectedItemId, nextItems));
			return;
		}
		setFileTree(nextItems);
	}, [data, selectedItemId]);

		const selectedItem = findItemInTree(fileTree, selectedItemId) ?? selectedItemSnapshot;

	const handleSelectItem = useCallback(
		(itemId: number | null) => {
			setSelectedItemId(itemId);
			setSelectedItemSnapshot(null);
			if (!itemId) return;
			if (expandedItems.includes(itemId)) {
				setExpandedItems((prev) => prev.filter((id) => id !== itemId));
			} else {
				setExpandedItems((prev) => [...prev, itemId]);
			}
		},
		[expandedItems],
	);

	const selectResolvedItem = useCallback((item: FileData | null) => {
		setSelectedItemSnapshot(item);
		setSelectedItemId(item?.id ?? null);
	}, []);

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
				selectedItem,
				handleSelectItem,
				selectResolvedItem,
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
				selectedItem,
				handleSelectItem,
				selectResolvedItem,
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
