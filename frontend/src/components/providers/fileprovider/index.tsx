import { apiFile } from '@/service'
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { FileContextProvider, FileContextType, FileData } from './fileContext'

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
	const [page, setPage] = useState<number>(1);
	const [selectedItem, setSelectedItem] = useState<FileData | null>(null);
	const [selectedIndex, setSelectedIndex] = useState<number | null>(null);

	const { status, data } = useQuery({
		queryKey: ['files'],
		queryFn: async (): Promise<PaginationResponse> => {
            const response = await apiFile.get<PaginationResponse>(`/`, {
                params: {
                    page: page,
                    page_size: pageSize,
                }
            });

			return response.data;
		},
	});

	const contextValue: FileContextType = {
		files: data?.items || [],
		status: status,
		selectedItem,
		selectedIndex,
		setSelectedItem,
		setSelectedIndex,
	};
	return <FileContextProvider value={contextValue}>{children}</FileContextProvider>;
};

export default FileProvider;
