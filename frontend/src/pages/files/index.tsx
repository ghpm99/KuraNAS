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
				<div style={{ gridArea: 'item-header' }}><ActionBar /></div>
				<div style={{ gridArea: 'tabs' }}><Tabs /></div>
				<FileContent />
				<div style={{ gridArea: 'sidebar' }}><FileDetails /></div>
			</div>
		</FilesLayout>
	);
};

export default FilePage;
