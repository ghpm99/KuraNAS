import { createContext, useContext } from 'react';

export type FileData = {
	id: number;
	name: string;
	path: string;
	type: number;
	format: string;
	size: number;
	updated_at: string;
	created_at: string;
	deleted_at: string;
	last_interaction: string;
	last_backup: string;
	check_sum: string;
	directory_content_count: number;
	file_children: FileData[];
};

export type RecentAccessFile = {
	id: number;
	ip_address: string;
	file_id: number;
	accessed_at: string;
};

export type FileContextType = {
	files: FileData[];
	recentAccessFiles: RecentAccessFile[];
	isLoadingAccessData: boolean;
	status: string;
	selectedItem: FileData | null;
	handleSelectItem: (itemId: number | null) => void;
	expandedItems: number[];
	fileListFilter: FileListFilterType;
	setFileListFilter: (filter: FileListFilterType) => void;
};

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

export type FileListFilterType = 'all' | 'recent' | 'starred';

const FileContext = createContext<FileContextType | undefined>(undefined);

export const FileContextProvider = FileContext.Provider;

export const useFile = () => {
	const context = useContext(FileContext);
	if (!context) {
		throw new Error('useFile must be used within a FileProvider');
	}

	return context;
};

export default useFile;
