import './App.css'

import ActionBar from '@/components/actionbar'
import FileCard from '@/components/filecard'
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
					<div className='file-grid'>
						<FileCard title='Q4 Sales Deck' metadata='Shared folder • 8 presentations' thumbnail='/placeholder.svg' />
						<FileCard title='Product Videos' metadata='Shared folder • 5 videos' thumbnail='/placeholder.svg' />
						<FileCard title='ROI Calculator' metadata='Shared file • 1 Excel' thumbnail='/placeholder.svg' />
					</div>
				</div>
			</div>
		</div>
	);
}
