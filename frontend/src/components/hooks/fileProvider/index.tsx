import { apiBase } from '@/service';

import { FileType } from '@/utils';
import { useInfiniteQuery, useMutation, useQuery } from '@tanstack/react-query';
import { useCallback, useEffect, useMemo, useState } from 'react';
import {
	FileContextProvider,
	FileContextType,
	FileData,
	FileListCategoryType,
	PaginationResponse,
	RecentAccessFile,
} from './fileContext';

const pageSize = 200;

const findItemInTree = (data: FileData[], itemId: number | null): FileData | null => {
	if (!itemId) return null;
	for (const item of data) {
		if (item.id === itemId) {
			return item;
		}
		if (item?.file_children?.length > 0) {
			const itemChildren = findItemInTree(item?.file_children, itemId);
			if (itemChildren) {
				return itemChildren;
			}
		}
	}

	return null;
};

const FileProvider = ({ children }: { children: React.ReactNode }) => {
	const [selectedItemId, setSelectedItemId] = useState<number | null>(null);
	const [selectedItem, setSelectedItem] = useState<FileData | null>(null);
	const [fileTree, setFileTree] = useState<FileData[]>([]);
	const [expandedItems, setExpandedItems] = useState<number[]>([]);
	const [fileListFilter, setFileListFilter] = useState<FileListCategoryType>('all');

	const queryParams = useMemo(
		() => ({
			page_size: pageSize,
			file_parent: selectedItemId ?? undefined,
		}),
		[selectedItemId]
	);

	const { status, data, refetch } = useInfiniteQuery({
		queryKey: ['files', queryParams, fileListFilter],
		queryFn: async ({ pageParam = 1 }): Promise<PaginationResponse> => {
			const response = await apiBase.get<PaginationResponse>(`/files/tree`, {
				params: { ...queryParams, page: pageParam, category: fileListFilter },
			});
			return response.data;
		},
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
		queryKey: ['filesRecent', 'tree', selectedItem],
		queryFn: async () => {
			if (selectedItem?.type !== FileType.File) return [];

			const response = await apiBase.get<RecentAccessFile[]>(`/files/recent/${selectedItemId}`);
			return response.data;
		},
		staleTime: 0,
	});

	const { mutate: updateStarredFile } = useMutation({
		mutationFn: async (itemId: number) => {
			await apiBase.post(`/files/starred/${itemId}`);
		},
		onSuccess: () => {
			refetch();
		},
	});

	const findAndAddChildren = useCallback((tree: FileData[], parentId: number, children: FileData[]): FileData[] => {
		return tree.map((node) => {
			if (node.id === parentId) {
				return { ...node, file_children: children };
			}
			if (node.file_children) {
				return { ...node, file_children: findAndAddChildren(node.file_children, parentId, children) };
			}
			return node;
		});
	}, []);

	useEffect(() => {
		if (!data) return;

		if (selectedItemId) {
			setFileTree((currentTree) => {
				const updatedTree = findAndAddChildren(currentTree, selectedItemId, data.pages[0].items);
				return updatedTree;
			});
		} else {
			setFileTree(data.pages[0].items);
		}
	}, [data, selectedItemId, findAndAddChildren]);

	useEffect(() => {
		if (selectedItemId) {
			const item = findItemInTree(fileTree, selectedItemId);
			setSelectedItem(item);
		} else {
			setSelectedItem(null);
		}
	}, [fileTree, selectedItemId]);

	const handleSelectItem = useCallback(
		(itemId: number | null) => {
			setSelectedItemId(itemId);
			if (!itemId) return;
			if (expandedItems.includes(itemId)) {
				setExpandedItems((prev) => prev.filter((id) => id !== itemId));
			} else {
				setExpandedItems((prev) => [...prev, itemId]);
			}
		},
		[expandedItems]
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
			selectedItem,
			handleSelectItem,
			expandedItems,
			recentAccessFiles: fileAccessData || [],
			isLoadingAccessData: isLoadingAccessData,
			fileListFilter,
			setFileListFilter,
			handleStarredItem,
		}),
		[
			fileTree,
			status,
			selectedItem,
			handleSelectItem,
			expandedItems,
			fileAccessData,
			isLoadingAccessData,
			fileListFilter,
			handleStarredItem,
		]
	);
	return <FileContextProvider value={contextValue}>{children}</FileContextProvider>;
};

export default FileProvider;
