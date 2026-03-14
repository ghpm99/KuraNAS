import FavoritesScreen from '@/components/favorites/FavoritesScreen';
import FileListFilterSync from '@/components/files/FileListFilterSync';
import FilesLayout from '@/components/files/filesLayout';

const FavoritesPage = () => {
	return (
		<FilesLayout>
			<>
				<FileListFilterSync filter='starred' />
				<FavoritesScreen />
			</>
		</FilesLayout>
	);
};

export default FavoritesPage;
