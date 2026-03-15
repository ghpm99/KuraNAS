import { IMusicData } from '@/components/providers/musicProvider/musicProvider';
import { Pagination } from '@/types/pagination';

export const MUSIC_COLLECTION_PAGE_SIZE = 50;

export const handleKeyboardActivation = (
	event: React.KeyboardEvent<HTMLElement>,
	onActivate: () => void,
) => {
	if (event.key === 'Enter' || event.key === ' ') {
		event.preventDefault();
		onActivate();
	}
};

export const shuffleTracks = (tracks: IMusicData[]) => [...tracks].sort(() => Math.random() - 0.5);

export const getFolderName = (path: string) => {
	const parts = path.split('/').filter(Boolean);
	return parts[parts.length - 1] || path;
};

export const loadAllTracks = async (
	fetchPage: (page: number, pageSize: number) => Promise<Pagination<IMusicData>>,
	pageSize = 200,
) => {
	const items: IMusicData[] = [];
	let page = 1;
	let hasNext = true;

	while (hasNext) {
		const response = await fetchPage(page, pageSize);
		items.push(...response.items);
		hasNext = response.pagination.has_next;
		page += 1;
	}

	return items;
};
