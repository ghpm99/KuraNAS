import { createContext, useContext } from 'react';

export type FileData = {
    id: number;
    name: string;
    path: string;
    parent_path: string;
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
    starred: boolean;
    file_children?: FileData[];
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
    handleSelectItem: (item: FileData | null) => void;
    handleStarredItem: (itemId: number) => void;
    uploadFiles: (files: FileList, targetPath?: string) => Promise<void>;
    createFolder: (name: string, parentPath?: string) => Promise<void>;
    movePath: (sourcePath: string, destinationPath: string) => Promise<void>;
    copyPath: (sourcePath: string, destinationPath: string) => Promise<void>;
    renamePath: (sourcePath: string, newName: string) => Promise<void>;
    deletePath: (path: string) => Promise<void>;
    rescanFiles: () => Promise<void>;
    expandedItems: number[];
    fileListFilter: FileListCategoryType;
    setFileListFilter: (filter: FileListCategoryType) => void;
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

export type FileListCategoryType = 'all' | 'recent' | 'starred';

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
