import ActionBar from '@/components/actionBar';
import FileContent from '@/components/fileContent';
import FileDetails from '@/components/fileDetails';
import Tabs from '@/components/tabs';
import './files.css';

const FilePage = () => {
	return (
		<div className='content'>
			<ActionBar />
			<Tabs />
			<FileContent />
			<FileDetails />
		</div>
	);
};

export default FilePage;
