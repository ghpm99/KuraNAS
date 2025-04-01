import { useQuery, useQueryClient } from '@tanstack/react-query'
import './App.css'

import ActionBar from '@/components/actionbar'
import FileCard from '@/components/filecard'
import Header from '@/components/header'
import Sidebar from '@/components/sidebar'
import Tabs from '@/components/tabs'

export type FileData = {
	id: number;
	name: string;
	format: string;
	size: number;
	updateAt: string;
	createdAt: string;
	lastInteraction: string;
	lastBackup: string;
};

export type Pagination = {
	hasNext: boolean;
	hasPrevious: boolean;
	page: number;
	pageSize: number;
};

export type PaginationResponse = {
	items: FileData[];
	pagination: Pagination;
};

function usePosts() {
	return useQuery({
		queryKey: ['files'],
		queryFn: async (): Promise<PaginationResponse> => {
			const response = await fetch('http://localhost:8080/api/v1/files/');
			return await response.json();
		},
	});
}

export default function NetflixStyleGallery() {
	const queryClient = useQueryClient();
	const { status, data, error, isFetching, isPending } = usePosts();
	// Sample categories and media items
	console.log(data);

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
