import { useInfiniteQuery } from '@tanstack/react-query';
import { IVideoData } from '@/types/video';

const fetchVideos = async ({ pageParam = 1 }): Promise<{ data: IVideoData[]; nextPage?: number }> => {
	const response = await fetch(`${import.meta.env.VITE_API_URL}/api/v1/files/videos?page=${pageParam}&page_size=15`);
	
	if (!response.ok) {
		throw new Error('Failed to fetch videos');
	}
	
	const result = await response.json();
	
	return {
		data: result.items || [],
		nextPage: result.pagination?.has_next ? pageParam + 1 : undefined,
	};
};

export const useVideos = () => {
	return useInfiniteQuery({
		queryKey: ['videos'],
		queryFn: fetchVideos,
		getNextPageParam: (lastPage) => lastPage.nextPage,
		initialPageParam: 1,
	});
};
