import { apiFile } from '@/service';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { FileContextProvider, FileContextType, FileData } from './fileContext';
import { useInfiniteQuery } from 'react-query'

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
	const [selectedItem, setSelectedItem] = useState<FileData | null>(null);
	const [fileTree, setFileTree] = useState<FileData[]>([]);

	console.log('selectedItem', selectedItem);

	const queryParams = useMemo(
		() => ({
			page_size: pageSize,
			path: selectedItem ? `${selectedItem?.path || ''}${selectedItem?.name || ''}` : undefined,
		}),
		[selectedItem]
	);

	const { status, data } = useInfiniteQuery({
		queryKey: ['files', queryParams],
		queryFn: async ({ pageParam = 1 }): Promise<PaginationResponse> => {
			const response = await apiFile.get<PaginationResponse>(`/`, {
				params: { ...queryParams, page: pageParam },
			});
			return response.data;
		},
		getNextPageParam: (lastPage) => {
			if (lastPage.pagination.hasNext) {
				return lastPage.pagination.page + 1;
			}
			return undefined;
		},
	});

	const findAndAddChildren = useCallback((tree: FileData[], parent: FileData, children: FileData[]): FileData[] => {
		return tree.map((node) => {
			if (node.id === parent.id) {
				return { ...node, file_children: children };
			}
			if (node.file_children) {
				return { ...node, file_children: findAndAddChildren(node.file_children, parent, children) };
			}
			return node;
		});
	}, []);

	useEffect(() => {
		if (!data) return;

		if (selectedItem) {
			setFileTree((currentTree) => {
				// Encontre o item pai na árvore e adicione os filhos
				const updatedTree = findAndAddChildren(currentTree, selectedItem, data.pages[0].items);
				return updatedTree;
			});
		} else {
			setFileTree(data.pages[0].items);
		}
	}, [data, selectedItem, findAndAddChildren]);

	const handleSelectItem = useCallback((item: FileData | null) => {
		setSelectedItem(item);
	}, []);

	const contextValue: FileContextType = useMemo(
		() => ({
			files: fileTree || [],
			status: status,
			selectedItem,
			handleSelectItem,
		}),
		[fileTree, status, selectedItem, handleSelectItem]
	);
	return <FileContextProvider value={contextValue}>{children}</FileContextProvider>;
};

export default FileProvider;
