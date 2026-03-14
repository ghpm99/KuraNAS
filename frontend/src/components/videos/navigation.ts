import { getVideoRoute, type VideoSection } from '@/app/routes';

export type VideoNavigationItem = {
	key: VideoSection;
	labelKey: string;
	descriptionKey: string;
};

export const videoNavigationItems: VideoNavigationItem[] = [
	{
		key: 'home',
		labelKey: 'VIDEO_SECTION_HOME',
		descriptionKey: 'VIDEO_SECTION_HOME_DESCRIPTION',
	},
	{
		key: 'continue',
		labelKey: 'VIDEO_SECTION_CONTINUE',
		descriptionKey: 'VIDEO_SECTION_CONTINUE_DESCRIPTION',
	},
	{
		key: 'series',
		labelKey: 'VIDEO_SECTION_SERIES',
		descriptionKey: 'VIDEO_SECTION_SERIES_DESCRIPTION',
	},
	{
		key: 'movies',
		labelKey: 'VIDEO_SECTION_MOVIES',
		descriptionKey: 'VIDEO_SECTION_MOVIES_DESCRIPTION',
	},
	{
		key: 'personal',
		labelKey: 'VIDEO_SECTION_PERSONAL',
		descriptionKey: 'VIDEO_SECTION_PERSONAL_DESCRIPTION',
	},
	{
		key: 'clips',
		labelKey: 'VIDEO_SECTION_CLIPS',
		descriptionKey: 'VIDEO_SECTION_CLIPS_DESCRIPTION',
	},
	{
		key: 'folders',
		labelKey: 'VIDEO_SECTION_FOLDERS',
		descriptionKey: 'VIDEO_SECTION_FOLDERS_DESCRIPTION',
	},
];

const videoRouteEntries = videoNavigationItems.map((item) => [getVideoRoute(item.key), item.key] as const);

export const getVideoSectionFromPath = (pathname: string): VideoSection => {
	const matchedEntry = videoRouteEntries.find(([route]) => route === pathname);

	return matchedEntry?.[1] ?? 'home';
};

export const getVideoSectionMeta = (section: VideoSection) => {
	const matchedItem = videoNavigationItems.find((item) => item.key === section);

	if (matchedItem) {
		return matchedItem;
	}

	return videoNavigationItems[0]!;
};
