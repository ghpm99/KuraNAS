import FavoritesScreen from '@/components/favorites/FavoritesScreen';
import FileListFilterSync from '@/features/files/files/FileListFilterSync';
import FilesLayout from '@/features/files/files/filesLayout';

const FavoritesPage = () => {
    return (
        <FilesLayout>
            <>
                <FileListFilterSync filter="starred" />
                <FavoritesScreen />
            </>
        </FilesLayout>
    );
};

export default FavoritesPage;
