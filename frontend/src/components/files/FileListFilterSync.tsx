import { useEffect } from 'react';
import useFile, { FileListCategoryType } from '@/components/providers/fileProvider/fileContext';

interface FileListFilterSyncProps {
    filter: FileListCategoryType;
}

export default function FileListFilterSync({ filter }: FileListFilterSyncProps) {
    const { fileListFilter, setFileListFilter } = useFile();

    useEffect(() => {
        if (fileListFilter !== filter) {
            setFileListFilter(filter);
        }
    }, [fileListFilter, filter, setFileListFilter]);

    return null;
}
