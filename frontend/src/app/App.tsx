import './App.css';

import ActionBar from '@/components/actionBar';
import FileContent from '@/components/fileContent';
import FileDetails from '@/components/fileDetails';
import Header from '@/components/header';
import Sidebar from '@/components/sidebar';
import Tabs from '@/components/tabs';

export default function App() {
	return (
		<>
			<div className='sidebar-header'>
				<h1 className='app-title'>KuraNAS</h1>
			</div>
			<Sidebar />
			<Header />
			<ActionBar />
			<Tabs />
			<FileContent />
			<FileDetails />
		</>
	);
}
