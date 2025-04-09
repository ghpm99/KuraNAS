import './App.css'

import ActionBar from '@/components/actionbar'
import FileContent from '@/components/filecontent'
import Header from '@/components/header'
import Sidebar from '@/components/sidebar'
import Tabs from '@/components/tabs'


export default function NetflixStyleGallery() {

	return (
		<div className='file-manager'>
			<Sidebar />
			<div className='main-content'>
				<Header />
				<div className='content'>
					<ActionBar />
					<Tabs />
					<FileContent />
				</div>
			</div>
		</div>
	);
}
