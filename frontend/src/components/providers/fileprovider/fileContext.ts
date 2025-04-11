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
	file_children: FileData[];
};

export type FileContextType = {
	files: FileData[];
	status: string;
	selectedItem: FileData | null;
	handleSelectItem: (item: FileData | null) => void;
	expandedItems: number[];
};

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
