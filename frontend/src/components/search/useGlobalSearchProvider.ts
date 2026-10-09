import { useEffect, useMemo, useState, type KeyboardEvent as ReactKeyboardEvent } from 'react';
import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { appRoutes, getAnalyticsRoute, getMusicRoute, getVideoRoute } from '@/app/routes';
import {
    getAlbumSearchRoute,
    getArtistSearchRoute,
    getFileSearchRoute,
    getFilesQuerySearchRoute,
    getImageSearchRoute,
    getPlaylistSearchRoute,
} from './searchResultRoutes';
import useI18n from '@/components/i18n/provider/i18nContext';
import useDebouncedValue from '@/components/hooks/useDebouncedValue/useDebouncedValue';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { useOptionalGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import {
    buildPlayableTrack,
    buildTrackPlaybackContext,
    formatTrackDuration,
} from './searchTrackPlayback';
import { searchGlobal, searchGlobalWithAI } from '@/service/search';
import { useLocation, useNavigate } from 'react-router-dom';

export type SearchItemKind =
    | 'action'
    | 'file'
    | 'folder'
    | 'artist'
    | 'album'
    | 'track'
    | 'playlist'
    | 'video'
    | 'image';

export type SearchDialogItem = {
    id: string;
    kind: SearchItemKind;
    label: string;
    description: string;
    meta?: string;
    keepsDialogOpen?: boolean;
    secondaryAction?: { label: string; onSelect: () => void };
    onSelect: () => void;
};

export type SearchDialogSection = {
    id: string;
    title: string;
    items: SearchDialogItem[];
};

const searchResultLimit = 6;
const aiSearchMinWords = 2;
const searchMinCharacters = 2;
const searchDebounceMs = 250;

const countWords = (value: string) => value.split(/\s+/).filter(Boolean).length;

const normalizeForMatching = (value: string) =>
    value
        .normalize('NFD')
        .replace(/[\u0300-\u036f]/g, '')
        .toLowerCase();

const isActionMatchingQuery = (query: string, label: string) => {
    const queryWords = normalizeForMatching(query).split(/\s+/).filter(Boolean);
    if (queryWords.length === 0) {
        return true;
    }

    const labelWords = normalizeForMatching(label).split(/\s+/).filter(Boolean);
    return queryWords.every((queryWord) =>
        labelWords.some((labelWord) => labelWord.startsWith(queryWord))
    );
};

export const useGlobalSearchProvider = () => {
    const { t } = useI18n();
    const navigate = useNavigate();
    const location = useLocation();
    const globalMusic = useOptionalGlobalMusic();
    const replaceQueue = globalMusic?.replaceQueue;
    const [open, setOpen] = useState(false);
    const [query, setQuery] = useState('');
    const [activeIndex, setActiveIndex] = useState(0);
    const [aiRequestedQuery, setAiRequestedQuery] = useState('');
    const debouncedQuery = useDebouncedValue(query, searchDebounceMs);
    const normalizedQuery = debouncedQuery.trim();
    const isDebouncing = query.trim() !== normalizedQuery;
    const hasSearchableQuery = normalizedQuery.length >= searchMinCharacters;

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
        [navigate, t]
    );

    const {
        data: baseQueryData,
        isFetching: isBaseFetching,
        isPlaceholderData: isBasePlaceholderData,
        error: baseError,
        refetch: refetchBase,
    } = useQuery({
        queryKey: ['global-search', normalizedQuery],
        queryFn: ({ signal }) => searchGlobal(normalizedQuery, searchResultLimit, signal),
        enabled: open && hasSearchableQuery,
        placeholderData: keepPreviousData,
    });

    const isAiRequested = aiRequestedQuery !== '' && aiRequestedQuery === normalizedQuery;
    const {
        data: aiData,
        isFetching: isAiFetching,
        error: aiError,
        refetch: refetchAi,
    } = useQuery({
        queryKey: ['global-search-ai', normalizedQuery],
        queryFn: ({ signal }) => searchGlobalWithAI(normalizedQuery, searchResultLimit, signal),
        enabled: open && isAiRequested,
        staleTime: 10 * 60 * 1000,
    });

    const baseData = hasSearchableQuery ? baseQueryData : undefined;
    const data = aiData ?? baseData;
    const isFetching = isBaseFetching || isAiFetching;
    const isUpdating = Boolean(data) && (isFetching || isDebouncing || isBasePlaceholderData);
    const searchError = (isAiRequested ? aiError : null) ?? (hasSearchableQuery ? baseError : null);
    const hasSearchError = Boolean(searchError) && !isFetching;
    const searchErrorMessage = hasSearchError
        ? (extractBackendErrorMessage(searchError) ?? '')
        : '';
    const suggestion = aiData?.suggestion ?? '';
    const canOfferAiSearch =
        Boolean(baseData) &&
        !aiData &&
        !isAiFetching &&
        countWords(normalizedQuery) >= aiSearchMinWords;

    const retrySearch = () => {
        if (isAiRequested && aiError) {
            void refetchAi();
            return;
        }
        void refetchBase();
    };

    const sections = useMemo<SearchDialogSection[]>(() => {
        const nextSections: SearchDialogSection[] = [];
        const appendActionsSection = () => {
            const matchingActions = quickActions.filter((action) =>
                isActionMatchingQuery(normalizedQuery, action.label)
            );
            if (matchingActions.length === 0) {
                return;
            }
            nextSections.push({
                id: 'actions',
                title: t('GLOBAL_SEARCH_SECTION_ACTIONS'),
                items: matchingActions,
            });
        };

        if (!data) {
            appendActionsSection();
            return nextSections;
        }

        const files = data.files.map<SearchDialogItem>((item) => ({
            id: `file-${item.id}`,
            kind: 'file',
            label: item.name,
            description: item.path,
            meta: item.format,
            onSelect: () => navigate(getFileSearchRoute(item.path)),
        }));
        if (files.length > 0) {
            nextSections.push({
                id: 'files',
                title: t('GLOBAL_SEARCH_SECTION_FILES'),
                items: files,
            });
        }

        const tracks = (data.tracks ?? []).map<SearchDialogItem>((item) => {
            const albumRoute = item.album_key ? getAlbumSearchRoute(item.album_key) : undefined;
            return {
                id: `track-${item.file_id}`,
                kind: 'track',
                label: item.title,
                description: [item.artist, item.album].filter(Boolean).join(' · ') || item.path,
                meta: formatTrackDuration(item.duration),
                secondaryAction: albumRoute
                    ? {
                          label: t('GLOBAL_SEARCH_OPEN_ALBUM'),
                          onSelect: () => navigate(albumRoute),
                      }
                    : undefined,
                onSelect: () => {
                    if (!replaceQueue) {
                        navigate(getFileSearchRoute(item.path));
                        return;
                    }
                    replaceQueue([buildPlayableTrack(item)], 0, buildTrackPlaybackContext(item));
                },
            };
        });
        if (tracks.length > 0) {
            nextSections.push({
                id: 'tracks',
                title: t('GLOBAL_SEARCH_SECTION_TRACKS'),
                items: tracks,
            });
        }

        const folders = data.folders.map<SearchDialogItem>((item) => ({
            id: `folder-${item.id}`,
            kind: 'folder',
            label: item.name,
            description: item.path,
            onSelect: () => navigate(getFileSearchRoute(item.path)),
        }));
        if (folders.length > 0) {
            nextSections.push({
                id: 'folders',
                title: t('GLOBAL_SEARCH_SECTION_FOLDERS'),
                items: folders,
            });
        }

        const images = data.images.map<SearchDialogItem>((item) => ({
            id: `image-${item.id}`,
            kind: 'image',
            label: item.name,
            description: item.path,
            meta: item.context || item.category,
            onSelect: () => navigate(getImageSearchRoute(item.id, item.path)),
        }));
        if (images.length > 0) {
            nextSections.push({
                id: 'images',
                title: t('GLOBAL_SEARCH_SECTION_IMAGES'),
                items: images,
            });
        }

        const videos = data.videos.map<SearchDialogItem>((item) => ({
            id: `video-${item.id}`,
            kind: 'video',
            label: item.name,
            description: item.path,
            meta: item.format,
            onSelect: () =>
                navigate(`/video/${item.id}`, {
                    state: { from: currentRoute, playlistId: null },
                }),
        }));
        if (videos.length > 0) {
            nextSections.push({
                id: 'videos',
                title: t('GLOBAL_SEARCH_SECTION_VIDEOS'),
                items: videos,
            });
        }

        const artists = data.artists.map<SearchDialogItem>((item) => ({
            id: `artist-${item.key}`,
            kind: 'artist',
            label: item.artist,
            description: t('GLOBAL_SEARCH_ARTIST_META', {
                tracks: String(item.track_count),
                albums: String(item.album_count),
            }),
            onSelect: () => navigate(getArtistSearchRoute(item.key)),
        }));
        if (artists.length > 0) {
            nextSections.push({
                id: 'artists',
                title: t('GLOBAL_SEARCH_SECTION_ARTISTS'),
                items: artists,
            });
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
            onSelect: () => navigate(getAlbumSearchRoute(item.key)),
        }));
        if (albums.length > 0) {
            nextSections.push({
                id: 'albums',
                title: t('GLOBAL_SEARCH_SECTION_ALBUMS'),
                items: albums,
            });
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
            onSelect: () => navigate(getPlaylistSearchRoute(item)),
        }));
        if (playlists.length > 0) {
            nextSections.push({
                id: 'playlists',
                title: t('GLOBAL_SEARCH_SECTION_PLAYLISTS'),
                items: playlists,
            });
        }

        if (files.length > 0 || folders.length > 0) {
            nextSections.push({
                id: 'files-see-all',
                title: t('GLOBAL_SEARCH_SECTION_MORE'),
                items: [
                    {
                        id: 'files-see-all-results',
                        kind: 'action',
                        label: t('GLOBAL_SEARCH_SEE_ALL_FILES'),
                        description: t('GLOBAL_SEARCH_SEE_ALL_FILES_DESCRIPTION', {
                            query: normalizedQuery,
                        }),
                        onSelect: () => navigate(getFilesQuerySearchRoute(normalizedQuery)),
                    },
                ],
            });
        }

        if (canOfferAiSearch) {
            nextSections.push({
                id: 'ai-search',
                title: t('GLOBAL_SEARCH_SECTION_MORE'),
                items: [
                    {
                        id: 'ai-search-expand',
                        kind: 'action',
                        label: t('GLOBAL_SEARCH_WITH_AI'),
                        description: t('GLOBAL_SEARCH_WITH_AI_DESCRIPTION', {
                            query: normalizedQuery,
                        }),
                        keepsDialogOpen: true,
                        onSelect: () => setAiRequestedQuery(normalizedQuery),
                    },
                ],
            });
        }

        appendActionsSection();
        return nextSections;
    }, [canOfferAiSearch, currentRoute, data, navigate, normalizedQuery, quickActions, replaceQueue, t]);

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
        if (!item.keepsDialogOpen) {
            closeSearch();
        }
    };

    const activateSecondaryAction = (item: SearchDialogItem) => {
        if (!item.secondaryAction) {
            return;
        }
        item.secondaryAction.onSelect();
        closeSearch();
    };

    const handleInputKeyDown = (
        event: ReactKeyboardEvent<HTMLInputElement | HTMLTextAreaElement>
    ) => {
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
            setActiveIndex(
                (current) => (current - 1 + flattenedItems.length) % flattenedItems.length
            );
            return;
        }

        if (event.key === 'Enter') {
            event.preventDefault();
            const currentItem = flattenedItems[activeIndex];
            if (!currentItem) {
                return;
            }
            if (event.shiftKey) {
                activateSecondaryAction(currentItem);
                return;
            }
            activateItem(currentItem);
        }
    };

    return {
        open,
        query,
        sections,
        isFetching,
        isUpdating,
        hasSearchError,
        searchErrorMessage,
        retrySearch,
        suggestion,
        activeItemId,
        shortcut,
        showEmptyState:
            hasSearchableQuery &&
            !isFetching &&
            !isDebouncing &&
            !hasSearchError &&
            sections.length === 0,
        openSearch,
        closeSearch,
        setQuery: updateQuery,
        setActiveIndex,
        handleInputKeyDown,
        activateItem,
        activateSecondaryAction,
    };
};

export default useGlobalSearchProvider;
