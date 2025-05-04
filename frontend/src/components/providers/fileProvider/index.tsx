import { apiBase } from '@/service';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { FileContextProvider, FileContextType, FileData } from './fileContext';
import { useInfiniteQuery } from '@tanstack/react-query';

export type Pagination = {
	hasNext: boolean;
	hasPrevious: boolean;
	page: number;
	pageSize: number;
};

export type PaginationResponse = {
	items: FileData[];
	pagination: Pagination;
};

const pageSize = 200;

const FileProvider = ({ children }: { children: React.ReactNode }) => {
	const [selectedItemId, setSelectedItemId] = useState<number | null>(null);
	const [fileTree, setFileTree] = useState<FileData[]>([]);
	const [expandedItems, setExpandedItems] = useState<number[]>([]);

	const queryParams = useMemo(
		() => ({
			page_size: pageSize,
			file_parent: selectedItemId ?? undefined,
		}),
		[selectedItemId]
	);

	const { status, data } = useInfiniteQuery({
		queryKey: ['files', queryParams],
		queryFn: async ({ pageParam = 1 }): Promise<PaginationResponse> => {
			const response = await apiBase.get<PaginationResponse>(`/files/tree`, {
				params: { ...queryParams, page: pageParam },
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

	const selectedItem = findItemInTree(fileTree, selectedItemId);

	const contextValue: FileContextType = useMemo(
		() => ({
			files: fileTree || [],
			status: status,
			selectedItem,
			handleSelectItem,
			expandedItems,
		}),
		[fileTree, status, selectedItem, handleSelectItem, expandedItems]
	);
	return <FileContextProvider value={contextValue}>{children}</FileContextProvider>;
};

export default FileProvider;
