import './App.css';

import ActionBar from '@/components/actionbar';
import FileContent from '@/components/filecontent';
import Header from '@/components/header';
import Sidebar from '@/components/sidebar';
import Tabs from '@/components/tabs';
import { apiBase } from '@/service';
import { useQuery } from '@tanstack/react-query';

export default function App() {
	const { status, data } = useQuery({
		queryKey: ['configuration'],
		queryFn: async () => {
			const response = await apiBase.get(`/configuration/translation`);
			return response.data;
		},
	});
	console.log('status', status);
	console.log('data', data);
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
