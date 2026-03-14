import { getMusicRoute, type MusicSection } from '@/app/routes';

export type MusicNavigationItem = {
	key: MusicSection;
	labelKey: string;
	descriptionKey: string;
};

export const musicNavigationItems: MusicNavigationItem[] = [
	{
		key: 'home',
		labelKey: 'MUSIC_SECTION_HOME',
		descriptionKey: 'MUSIC_HOME_DESCRIPTION',
	},
	{
		key: 'playlists',
		labelKey: 'MUSIC_PLAYLISTS',
		descriptionKey: 'MUSIC_PLAYLISTS_DESCRIPTION',
	},
	{
		key: 'artists',
		labelKey: 'MUSIC_ARTISTS',
		descriptionKey: 'MUSIC_ARTISTS_DESCRIPTION',
	},
	{
		key: 'albums',
		labelKey: 'MUSIC_ALBUMS',
		descriptionKey: 'MUSIC_ALBUMS_DESCRIPTION',
	},
	{
		key: 'genres',
		labelKey: 'MUSIC_GENRES',
		descriptionKey: 'MUSIC_GENRES_DESCRIPTION',
	},
	{
		key: 'folders',
		labelKey: 'MUSIC_FOLDERS',
		descriptionKey: 'MUSIC_FOLDERS_DESCRIPTION',
	},
];

const musicRouteEntries = musicNavigationItems.map((item) => [getMusicRoute(item.key), item.key] as const);

export const getMusicSectionFromPath = (pathname: string): MusicSection => {
	const matchedEntry = musicRouteEntries.find(([route]) => route === pathname);

	return matchedEntry?.[1] ?? 'home';
};

export const getMusicSectionMeta = (section: MusicSection) => {
	const matchedItem = musicNavigationItems.find((item) => item.key === section);

	if (matchedItem) {
		return matchedItem;
	}

	return musicNavigationItems[0]!;
};
