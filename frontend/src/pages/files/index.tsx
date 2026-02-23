import ActionBar from '@/components/actionBar';
import FileContent from '@/components/fileContent';
import FileDetails from '@/components/fileDetails';
import Tabs from '@/components/tabs';
import './files.css';
import FilesLayout from '@/components/files/filesLayout';

const FilePage = () => {
	return (
		<FilesLayout>
			<div className='content'>
				<ActionBar />
				<Tabs />
				<FileContent />
				<FileDetails />
			</div>
		</FilesLayout>
	);
};

export default FilePage;
