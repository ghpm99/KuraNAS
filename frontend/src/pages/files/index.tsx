import FilesExplorerScreen from '@/features/files/files/FilesExplorerScreen';
import FileListFilterSync from '@/features/files/files/FileListFilterSync';
import FilesLayout from '@/features/files/files/filesLayout';

const FilePage = () => {
    return (
        <FilesLayout>
            <>
                <FileListFilterSync filter="all" />
                <FilesExplorerScreen />
            </>
        </FilesLayout>
    );
};

export default FilePage;
