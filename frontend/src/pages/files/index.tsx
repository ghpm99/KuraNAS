import { useLocation } from 'react-router-dom';
import FilesExplorerScreen from '@/features/files/files/FilesExplorerScreen';
import FileListFilterSync from '@/features/files/files/FileListFilterSync';
import FilesLayout from '@/features/files/files/filesLayout';
import { parseFileListFilter } from '@/features/files/files/fileListFilterFromSearch';

const FilePage = () => {
    const { search } = useLocation();

    return (
        <FilesLayout>
            <>
                <FileListFilterSync filter={parseFileListFilter(search)} />
                <FilesExplorerScreen />
            </>
        </FilesLayout>
    );
};

export default FilePage;
