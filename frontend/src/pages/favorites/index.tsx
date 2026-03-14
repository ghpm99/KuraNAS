import ActionBar from '@/components/actionBar';
import FileContent from '@/components/fileContent';
import FileDetails from '@/components/fileDetails';
import FileListFilterSync from '@/components/files/FileListFilterSync';
import FilesLayout from '@/components/files/filesLayout';
import '@/pages/files/files.css';

const FavoritesPage = () => {
	return (
		<FilesLayout>
			<div className='content'>
				<FileListFilterSync filter='starred' />
				<div style={{ gridArea: 'item-header' }}><ActionBar /></div>
				<div style={{ gridArea: 'tabs' }} />
				<FileContent />
				<div style={{ gridArea: 'sidebar' }}><FileDetails /></div>
			</div>
		</FilesLayout>
	);
};

export default FavoritesPage;
