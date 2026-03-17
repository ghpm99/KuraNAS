import FilesExplorerScreen from '@/components/files/FilesExplorerScreen';
import FileListFilterSync from '@/components/files/FileListFilterSync';
import FilesLayout from '@/components/files/filesLayout';

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
