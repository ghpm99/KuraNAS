import { useQuery } from '@tanstack/react-query';
import { getVideoHomeCatalog } from '@/service/videoPlayback';

export const useVideos = () => {
	return useQuery({
		queryKey: ['video-home-catalog'],
		queryFn: () => getVideoHomeCatalog(24),
	});
};
