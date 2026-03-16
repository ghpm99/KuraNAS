import { useDeferredValue, useEffect, useMemo, useState, type KeyboardEvent as ReactKeyboardEvent } from 'react';
import { useQuery } from '@tanstack/react-query';
import { appRoutes, getAnalyticsRoute, getMusicRoute, getVideoRoute } from '@/app/routes';
import { getVideoDetailRoute, getVideoSectionForPlaylist } from '@/components/videos/navigation';
import useI18n from '@/components/i18n/provider/i18nContext';
import { searchGlobal } from '@/service/search';
import { useLocation, useNavigate } from 'react-router-dom';

export type SearchItemKind = 'action' | 'file' | 'folder' | 'artist' | 'album' | 'playlist' | 'video' | 'image';

export type SearchDialogItem = {
	id: string;
	kind: SearchItemKind;
	label: string;
	description: string;
	meta?: string;
	onSelect: () => void;
};

export type SearchDialogSection = {
	id: string;
	title: string;
	items: SearchDialogItem[];
};

const searchResultLimit = 6;

const slugify = (value: string) =>
	value
		.normalize('NFD')
		.replace(/[\u0300-\u036f]/g, '')
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-+|-+$/g, '');

const matchesQuery = (query: string, ...values: Array<string | undefined>) => {
	const normalizedQuery = query.trim().toLowerCase();
	if (!normalizedQuery) {
		return true;
	}

	return values.some((value) => value?.toLowerCase().includes(normalizedQuery));
};

export const useGlobalSearchProvider = () => {
	const { t } = useI18n();
	const navigate = useNavigate();
	const location = useLocation();
	const [open, setOpen] = useState(false);
	const [query, setQuery] = useState('');
	const [activeIndex, setActiveIndex] = useState(0);
	const deferredQuery = useDeferredValue(query);
	const normalizedQuery = deferredQuery.trim();

	const shortcut = useMemo(() => {
		if (typeof window === 'undefined') {
			return 'Ctrl+K';
		}

		const platform = window.navigator.platform.toLowerCase();
		return platform.includes('mac') ? 'Cmd+K' : 'Ctrl+K';
	}, []);

	const currentRoute = `${location.pathname}${location.search}`;

	const quickActions = useMemo<SearchDialogItem[]>(
		() => [
			{
				id: 'action-home',
				kind: 'action',
				label: t('HOME'),
				description: t('HOME_PAGE_DESCRIPTION'),
				onSelect: () => navigate(appRoutes.home),
			},
			{
				id: 'action-files',
				kind: 'action',
				label: t('FILES'),
				description: t('FILES_PAGE_DESCRIPTION'),
				onSelect: () => navigate(appRoutes.files),
			},
			{
				id: 'action-favorites',
				kind: 'action',
				label: t('STARRED_FILES'),
				description: t('FAVORITES_PAGE_DESCRIPTION'),
				onSelect: () => navigate(appRoutes.favorites),
			},
			{
				id: 'action-images',
				kind: 'action',
				label: t('NAV_IMAGES'),
				description: t('IMAGES_SECTION_RECENT_DESCRIPTION'),
				onSelect: () => navigate(appRoutes.images),
			},
			{
				id: 'action-music',
				kind: 'action',
				label: t('NAV_MUSIC'),
				description: t('MUSIC_HOME_DESCRIPTION'),
				onSelect: () => navigate(appRoutes.music),
			},
			{
				id: 'action-music-artists',
				kind: 'action',
				label: t('MUSIC_ARTISTS'),
				description: t('MUSIC_ARTISTS_DESCRIPTION'),
				onSelect: () => navigate(getMusicRoute('artists')),
			},
			{
				id: 'action-music-albums',
				kind: 'action',
				label: t('MUSIC_ALBUMS'),
				description: t('MUSIC_ALBUMS_DESCRIPTION'),
				onSelect: () => navigate(getMusicRoute('albums')),
			},
			{
				id: 'action-music-playlists',
				kind: 'action',
				label: t('MUSIC_PLAYLISTS'),
				description: t('MUSIC_PLAYLISTS_DESCRIPTION'),
				onSelect: () => navigate(getMusicRoute('playlists')),
			},
			{
				id: 'action-videos',
				kind: 'action',
				label: t('NAV_VIDEOS'),
				description: t('VIDEO_SECTION_HOME_DESCRIPTION'),
				onSelect: () => navigate(appRoutes.videos),
			},
			{
				id: 'action-videos-continue',
				kind: 'action',
					label: t('VIDEO_SECTION_CONTINUE'),
					description: t('VIDEO_SECTION_CONTINUE_DESCRIPTION'),
					onSelect: () => navigate(getVideoRoute('continue')),
			},
			{
				id: 'action-analytics',
				kind: 'action',
				label: t('ANALYTICS'),
				description: t('ANALYTICS_SECTION_OVERVIEW_DESCRIPTION'),
				onSelect: () => navigate(appRoutes.analytics),
			},
			{
				id: 'action-analytics-library',
				kind: 'action',
				label: t('ANALYTICS_SECTION_LIBRARY'),
				description: t('ANALYTICS_SECTION_LIBRARY_DESCRIPTION'),
				onSelect: () => navigate(getAnalyticsRoute('library')),
			},
			{
				id: 'action-settings',
				kind: 'action',
				label: t('SETTINGS'),
				description: t('SETTINGS_PAGE_DESCRIPTION'),
				onSelect: () => navigate(appRoutes.settings),
			},
			{
				id: 'action-about',
				kind: 'action',
				label: t('ABOUT'),
				description: t('ABOUT_PAGE_DESCRIPTION'),
				onSelect: () => navigate(appRoutes.about),
			},
		],
		[navigate, t],
	);

	const { data, isFetching } = useQuery({
		queryKey: ['global-search', normalizedQuery],
		queryFn: () => searchGlobal(normalizedQuery, searchResultLimit),
		enabled: open && normalizedQuery.length >= 2,
	});

	const sections = useMemo<SearchDialogSection[]>(() => {
		const nextSections: SearchDialogSection[] = [];

		const filteredActions = quickActions.filter((action) => matchesQuery(normalizedQuery, action.label, action.description));
		if (filteredActions.length > 0) {
			nextSections.push({
				id: 'actions',
				title: t('GLOBAL_SEARCH_SECTION_ACTIONS'),
				items: filteredActions,
			});
		}

		if (!data) {
			return nextSections;
		}

		const files = data.files.map<SearchDialogItem>((item) => ({
			id: `file-${item.id}`,
			kind: 'file',
			label: item.name,
			description: item.path,
			meta: item.format,
			onSelect: () => navigate(`${appRoutes.files}${item.path}`),
		}));
		if (files.length > 0) {
			nextSections.push({ id: 'files', title: t('GLOBAL_SEARCH_SECTION_FILES'), items: files });
		}

		const folders = data.folders.map<SearchDialogItem>((item) => ({
			id: `folder-${item.id}`,
			kind: 'folder',
			label: item.name,
			description: item.path,
			onSelect: () => navigate(`${appRoutes.files}${item.path}`),
		}));
		if (folders.length > 0) {
			nextSections.push({ id: 'folders', title: t('GLOBAL_SEARCH_SECTION_FOLDERS'), items: folders });
		}

		const artists = data.artists.map<SearchDialogItem>((item) => ({
			id: `artist-${item.key}`,
			kind: 'artist',
			label: item.artist,
			description: t('GLOBAL_SEARCH_ARTIST_META', {
				tracks: String(item.track_count),
				albums: String(item.album_count),
			}),
				onSelect: () => navigate({ pathname: getMusicRoute('artists'), search: `?artist=${encodeURIComponent(item.key)}` }),
		}));
		if (artists.length > 0) {
			nextSections.push({ id: 'artists', title: t('GLOBAL_SEARCH_SECTION_ARTISTS'), items: artists });
		}

		const albums = data.albums.map<SearchDialogItem>((item) => ({
			id: `album-${item.key}`,
			kind: 'album',
			label: item.album,
			description: t('GLOBAL_SEARCH_ALBUM_META', {
				artist: item.artist,
				tracks: String(item.track_count),
			}),
			meta: item.year,
				onSelect: () => navigate({ pathname: getMusicRoute('albums'), search: `?album=${encodeURIComponent(item.key)}` }),
		}));
		if (albums.length > 0) {
			nextSections.push({ id: 'albums', title: t('GLOBAL_SEARCH_SECTION_ALBUMS'), items: albums });
		}

		const playlists = data.playlists.map<SearchDialogItem>((item) => ({
			id: `playlist-${item.scope}-${item.id}`,
			kind: 'playlist',
			label: item.name,
			description:
				item.scope === 'music'
					? t('GLOBAL_SEARCH_PLAYLIST_META', {
							scope: t('NAV_MUSIC'),
							count: String(item.count),
					  })
					: t('GLOBAL_SEARCH_PLAYLIST_META', {
							scope: t('NAV_VIDEOS'),
							count: String(item.count),
					  }),
			meta: item.scope === 'video' ? item.classification : item.description,
			onSelect: () => {
				if (item.scope === 'music') {
						navigate({ pathname: getMusicRoute('playlists'), search: `?playlist=${item.id}` });
						return;
					}

				const section = getVideoSectionForPlaylist({
					type: item.source_path ? item.description : 'custom',
					classification: item.classification,
				});
				navigate(getVideoDetailRoute(section, slugify(item.name)));
			},
		}));
		if (playlists.length > 0) {
			nextSections.push({ id: 'playlists', title: t('GLOBAL_SEARCH_SECTION_PLAYLISTS'), items: playlists });
		}

		const videos = data.videos.map<SearchDialogItem>((item) => ({
			id: `video-${item.id}`,
			kind: 'video',
			label: item.name,
			description: item.path,
			meta: item.format,
			onSelect: () => navigate(`/video/${item.id}`, { state: { from: currentRoute, playlistId: null } }),
		}));
		if (videos.length > 0) {
			nextSections.push({ id: 'videos', title: t('GLOBAL_SEARCH_SECTION_VIDEOS'), items: videos });
		}

		const images = data.images.map<SearchDialogItem>((item) => ({
			id: `image-${item.id}`,
			kind: 'image',
			label: item.name,
			description: item.path,
			meta: item.context || item.category,
				onSelect: () =>
					navigate({
						pathname: appRoutes.images,
						search: `?image=${item.id}&imagePath=${encodeURIComponent(item.path)}`,
					}),
			}));
		if (images.length > 0) {
			nextSections.push({ id: 'images', title: t('GLOBAL_SEARCH_SECTION_IMAGES'), items: images });
		}

		return nextSections;
	}, [currentRoute, data, navigate, normalizedQuery, quickActions, t]);

	const flattenedItems = useMemo(() => sections.flatMap((section) => section.items), [sections]);
	const activeItemId = flattenedItems[activeIndex]?.id ?? '';

	useEffect(() => {
		const handleKeyDown = (event: KeyboardEvent) => {
			if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
				event.preventDefault();
				setOpen((current) => !current);
			}
		};

		window.addEventListener('keydown', handleKeyDown);
		return () => window.removeEventListener('keydown', handleKeyDown);
	}, []);

	const closeSearch = () => {
		setOpen(false);
		setQuery('');
		setActiveIndex(0);
	};

	const openSearch = () => {
		setActiveIndex(0);
		setOpen(true);
	};

	const updateQuery = (value: string) => {
		setActiveIndex(0);
		setQuery(value);
	};

	const activateItem = (item: SearchDialogItem) => {
		item.onSelect();
		closeSearch();
	};

	const handleInputKeyDown = (event: ReactKeyboardEvent<HTMLInputElement | HTMLTextAreaElement>) => {
		if (flattenedItems.length === 0) {
			return;
		}

		if (event.key === 'ArrowDown') {
			event.preventDefault();
			setActiveIndex((current) => (current + 1) % flattenedItems.length);
			return;
		}

		if (event.key === 'ArrowUp') {
			event.preventDefault();
			setActiveIndex((current) => (current - 1 + flattenedItems.length) % flattenedItems.length);
			return;
		}

		if (event.key === 'Enter') {
			event.preventDefault();
			const currentItem = flattenedItems[activeIndex];
			if (currentItem) {
				activateItem(currentItem);
			}
		}
	};

	return {
		open,
		query,
		sections,
		isFetching,
		activeItemId,
		shortcut,
		showEmptyState: normalizedQuery.length >= 2 && !isFetching && sections.length === 0,
		openSearch,
		closeSearch,
		setQuery: updateQuery,
		setActiveIndex,
		handleInputKeyDown,
		activateItem,
	};
};

export default useGlobalSearchProvider;
