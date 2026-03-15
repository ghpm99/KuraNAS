import FilesExplorerScreen from '@/components/files/FilesExplorerScreen';
import FileListFilterSync from '@/components/files/FileListFilterSync';
import FilePathSync from '@/components/files/FilePathSync';
import FilesLayout from '@/components/files/filesLayout';

const FilePage = () => {
	return (
		<FilesLayout>
			<>
				<FileListFilterSync filter='all' />
				<FilePathSync />
				<FilesExplorerScreen />
			</>
		</FilesLayout>
	);
};

export default FilePage;
