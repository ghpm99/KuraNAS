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
	if (pathname === getVideoRoute('home')) {
		return 'home';
	}

	const matchedEntry = videoRouteEntries.find(([route, section]) => {
		if (section === 'home') {
			return false;
		}

		return pathname === route || pathname.startsWith(`${route}/`);
	});

	return matchedEntry?.[1] ?? 'home';
};

export const getVideoDetailSlugFromPath = (pathname: string) => {
	const segments = pathname.split('/').filter(Boolean);
	if (segments[0] !== 'videos' || segments.length < 3) {
		return '';
	}

	return decodeURIComponent(segments[2] ?? '');
};

export const getVideoDetailRoute = (section: Exclude<VideoSection, 'home'>, slug: string) =>
	`${getVideoRoute(section)}/${encodeURIComponent(slug)}`;

export const getVideoSectionMeta = (section: VideoSection) => {
	const matchedItem = videoNavigationItems.find((item) => item.key === section);

	if (matchedItem) {
		return matchedItem;
	}

	return videoNavigationItems[0]!;
};
