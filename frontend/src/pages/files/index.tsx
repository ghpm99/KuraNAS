import ActionBar from '@/components/actionBar';
import FileContent from '@/components/fileContent';
import FileDetails from '@/components/fileDetails';
import Tabs from '@/components/tabs';
import './files.css';
import FileProvider from '@/components/providers/fileProvider';
import Sidebar from '@/components/sidebar';
import FolderTree from '@/components/sidebar/components/folderTree';

const FilePage = () => {
	return (
		<FileProvider>
			<Sidebar>
				<FolderTree />
			</Sidebar>
			<div className='content'>
				<ActionBar />
				<Tabs />
				<FileContent />
				<FileDetails />
			</div>
		</FileProvider>
	);
};

export default FilePage;
