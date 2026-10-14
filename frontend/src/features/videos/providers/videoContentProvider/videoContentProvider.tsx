/* eslint-disable react-refresh/only-export-components */
import {
    addVideoToPlaylist,
    createVideoPlaylist,
    deleteVideoPlaylist,
    getVideoContinueWatching,
    getVideoHomeCatalog,
    getVideoLibraryFiles,
    getVideoPlaylistById,
    getVideoPlaylistItemsPage,
    getVideoPlaylists,
    setVideoPlaylistHidden,
    type VideoCatalogItemDto,
    type VideoContinueItemDto,
    reorderVideoPlaylist,
    removeVideoFromPlaylist,
    setVideoWatched,
    updateVideoPlaylistName,
    type VideoFileDto,
    type VideoLibraryFilters,
    type VideoLibrarySort,
    type VideoPlaylistDto,
    type VideoPlaylistItemDto,
} from '@/service/videoPlayback';
import {
    useHiddenVideoPlaylists,
    useVideoSectionPlaylists,
    videoQueryKeys,
} from './useVideoQueries';
import { noVideoLibraryFilters } from './videoLibraryFilters';
import { loadVideoLibrarySort, saveVideoLibrarySort } from './videoLibrarySortPreference';
import { toVideoQueryFailure, type VideoQueryFailure } from './videoQueryFailure';
import { type VideoSection } from '@/app/routes';
import {
    getVideoDetailRoute,
    getVideoDetailSlugFromPath,
    getVideoPlaylistIdFromSearch,
    getVideoSectionForPlaylist,
    getVideoSectionFromPath,
} from '@/features/videos/components/navigation';
import {
    useInfiniteQuery,
    useMutation,
    useQuery,
    useQueryClient,
    type InfiniteData,
} from '@tanstack/react-query';
import type { Pagination } from '@/types/pagination';
import { createContext, useContext, useMemo, useState, type ReactNode } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import type { VideoFeedbackAction } from '@/features/videos/components/videoContent/components/VideoFeedbackSnackbar';
import useFileSearchQuery from '@/features/files/search/useFileSearchQuery';
import useI18n from '@/components/i18n/provider/i18nContext';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';

const VIDEO_LIBRARY_PAGE_SIZE = 60;
const VIDEO_PLAYLIST_ITEMS_PAGE_SIZE = 50;
const VIDEO_SECTION_GRID_PAGE_SIZE = 24;
const VIDEO_HIDDEN_PLAYLISTS_PAGE_SIZE = 24;
const VIDEO_HOME_RAIL_PAGE_SIZE = 4;
const VIDEO_HOME_CATALOG_LIMIT = 12;
const VIDEO_CONTINUE_WATCHING_LIMIT = 24;

type PlaylistItemsPages = InfiniteData<Pagination<VideoPlaylistItemDto>>;

type FeedbackState = {
    open: boolean;
    message: string;
    severity: 'success' | 'error';
    action?: VideoFeedbackAction;
};

export interface VideoContentContextData {
    currentSection: VideoSection;
    playlists: VideoPlaylistDto[];
    allVideos: VideoFileDto[];
    filteredVideos: VideoFileDto[];
    continueWatchingItems: VideoContinueItemDto[];
    seriesPlaylists: VideoPlaylistDto[];
    moviePlaylists: VideoPlaylistDto[];
    personalPlaylists: VideoPlaylistDto[];
    clipPlaylists: VideoPlaylistDto[];
    folderPlaylists: VideoPlaylistDto[];
    hiddenPlaylists: VideoPlaylistDto[];
    recentCatalogItems: VideoCatalogItemDto[];
    selectedPlaylistSummary: VideoPlaylistDto | null;
    selectedPlaylistDetail: VideoPlaylistDto | null;
    isLoadingPlaylists: boolean;
    isLoadingVideos: boolean;
    isLoadingSelectedPlaylist: boolean;
    isFetchingMoreSelectedPlaylistItems: boolean;
    hasMoreSelectedPlaylistItems: boolean;
    isLoadingHomeCatalog: boolean;
    isLoadingContinueWatching: boolean;
    playlistsFailure: VideoQueryFailure | null;
    videosFailure: VideoQueryFailure | null;
    selectedPlaylistFailure: VideoQueryFailure | null;
    homeCatalogFailure: VideoQueryFailure | null;
    continueWatchingFailure: VideoQueryFailure | null;
    isFetchingMoreVideos: boolean;
    hasMoreVideos: boolean;
    isFetchingMoreSectionPlaylists: boolean;
    hasMoreSectionPlaylists: boolean;
    isFetchingMoreHiddenPlaylists: boolean;
    hasMoreHiddenPlaylists: boolean;
    isChangingPlaylistVisibility: boolean;
    isAddingToPlaylist: boolean;
    isCreatingPlaylist: boolean;
    isDeletingPlaylist: boolean;
    isRenamingPlaylist: boolean;
    isRemovingFromPlaylist: boolean;
    isReorderingPlaylist: boolean;
    videoSearch: string;
    librarySort: VideoLibrarySort;
    libraryFilters: VideoLibraryFilters;
    feedback: FeedbackState;
    setVideoSearch: (value: string) => void;
    setLibrarySort: (sort: VideoLibrarySort) => void;
    setLibraryFilters: (filters: VideoLibraryFilters) => void;
    closeFeedback: () => void;
    loadMoreVideos: () => void;
    loadMoreSectionPlaylists: () => void;
    loadMoreHiddenPlaylists: () => void;
    loadMoreSelectedPlaylistItems: () => void;
    selectPlaylist: (playlist: VideoPlaylistDto) => void;
    clearSelectedPlaylist: () => void;
    playVideo: (videoId: number, playlistId?: number | null) => void;
    openPlaylistVideo: (videoId: number) => void;
    addVideoFromLibrary: (videoId: number, playlistId: number) => void;
    createPlaylist: (name: string) => void;
    createPlaylistWithVideo: (name: string, videoId: number) => void;
    deleteSelectedPlaylist: () => void;
    hidePlaylist: (playlist: VideoPlaylistDto) => void;
    hideSelectedPlaylist: () => void;
    showPlaylist: (playlist: VideoPlaylistDto) => void;
    renameSelectedPlaylist: (name: string) => void;
    removeVideoFromSelectedPlaylist: (videoId: number) => void;
    moveSelectedPlaylistItem: (index: number, direction: -1 | 1) => void;
    setVideoWatched: (videoId: number, watched: boolean) => void;
}

const VideoContentContext = createContext<VideoContentContextData | undefined>(undefined);

const gridSections: VideoSection[] = ['series', 'personal', 'clips'];

const flattenSectionPages = (pages?: Pagination<VideoPlaylistDto>[]) =>
    pages?.flatMap((page) => page.items ?? []) ?? [];

const slugify = (value: string) =>
    value
        .normalize('NFD')
        .replace(/[\u0300-\u036f]/g, '')
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/^-+|-+$/g, '');

export function VideoContentProvider({ children }: { children: ReactNode }) {
    const { t } = useI18n();
    const navigate = useNavigate();
    const location = useLocation();
    const queryClient = useQueryClient();
    const {
        inputValue: videoSearch,
        setInputValue: setVideoSearch,
        activeQuery: settledVideoSearch,
    } = useFileSearchQuery({ minQueryLength: 1, shouldAlwaysReplaceHistory: true });
    const [librarySort, setLibrarySortState] = useState(loadVideoLibrarySort);
    const setLibrarySort = (nextSort: VideoLibrarySort) => {
        setLibrarySortState(nextSort);
        saveVideoLibrarySort(nextSort);
    };
    const [libraryFilters, setLibraryFilters] =
        useState<VideoLibraryFilters>(noVideoLibraryFilters);
    const [feedback, setFeedback] = useState<FeedbackState>({
        open: false,
        message: '',
        severity: 'success',
    });
    const currentSection = getVideoSectionFromPath(location.pathname);

    const playlistSlug = getVideoDetailSlugFromPath(location.pathname);
    const playlistIdFromSearch = getVideoPlaylistIdFromSearch(location.search);
    const isHome = currentSection === 'home';
    const sectionPageSize = isHome ? VIDEO_HOME_RAIL_PAGE_SIZE : VIDEO_SECTION_GRID_PAGE_SIZE;
    const isSectionListed = (section: VideoSection) =>
        isHome ? true : gridSections.includes(section) && currentSection === section;

    const seriesQuery = useVideoSectionPlaylists(
        'series',
        sectionPageSize,
        isSectionListed('series')
    );
    const moviesQuery = useVideoSectionPlaylists('movies', sectionPageSize, isHome);
    const personalQuery = useVideoSectionPlaylists(
        'personal',
        sectionPageSize,
        isSectionListed('personal')
    );
    const clipsQuery = useVideoSectionPlaylists('clips', sectionPageSize, isSectionListed('clips'));
    const foldersQuery = useVideoSectionPlaylists('folders', sectionPageSize, isHome);
    const hiddenQuery = useHiddenVideoPlaylists(
        VIDEO_HIDDEN_PLAYLISTS_PAGE_SIZE,
        currentSection === 'hidden'
    );

    const isFullPlaylistListNeeded =
        currentSection === 'folders' || Boolean(playlistSlug && !playlistIdFromSearch);
    const playlistsQuery = useQuery({
        queryKey: videoQueryKeys.playlists,
        queryFn: () => getVideoPlaylists(false),
        enabled: isFullPlaylistListNeeded,
    });
    const { data: playlists = [] } = playlistsQuery;
    const homeCatalogQuery = useQuery({
        queryKey: videoQueryKeys.homeCatalog,
        queryFn: () => getVideoHomeCatalog(VIDEO_HOME_CATALOG_LIMIT),
    });
    const { data: homeCatalog, isLoading: isLoadingHomeCatalog } = homeCatalogQuery;
    const videoLibraryQuery = useInfiniteQuery({
        queryKey: videoQueryKeys.libraryFiles(settledVideoSearch, librarySort, libraryFilters),
        queryFn: ({ pageParam = 1 }) =>
            getVideoLibraryFiles(
                pageParam,
                VIDEO_LIBRARY_PAGE_SIZE,
                settledVideoSearch,
                librarySort,
                libraryFilters
            ),
        initialPageParam: 1,
        getNextPageParam: (lastPage) =>
            lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined,
    });
    const {
        data: videoLibraryData,
        isLoading: isLoadingVideos,
        isFetchingNextPage: isFetchingMoreVideos,
        hasNextPage: hasMoreVideos = false,
        fetchNextPage,
    } = videoLibraryQuery;

    const continueWatchingQuery = useQuery({
        queryKey: videoQueryKeys.continueWatching,
        queryFn: () => getVideoContinueWatching(VIDEO_CONTINUE_WATCHING_LIMIT),
    });
    const { data: continueWatchingItems = [], isLoading: isLoadingContinueWatching } =
        continueWatchingQuery;

    const seriesPlaylists = useMemo(
        () => flattenSectionPages(seriesQuery.data?.pages),
        [seriesQuery.data]
    );
    const moviePlaylists = useMemo(
        () => flattenSectionPages(moviesQuery.data?.pages),
        [moviesQuery.data]
    );
    const personalPlaylists = useMemo(
        () => flattenSectionPages(personalQuery.data?.pages),
        [personalQuery.data]
    );
    const clipPlaylists = useMemo(
        () => flattenSectionPages(clipsQuery.data?.pages),
        [clipsQuery.data]
    );
    const folderPlaylists = useMemo(
        () => flattenSectionPages(foldersQuery.data?.pages),
        [foldersQuery.data]
    );

    const hiddenPlaylists = useMemo(
        () => flattenSectionPages(hiddenQuery.data?.pages),
        [hiddenQuery.data]
    );

    const activeListQueries = useMemo(() => {
        if (currentSection === 'hidden') return [hiddenQuery];
        if (isHome) return [seriesQuery, moviesQuery, personalQuery, clipsQuery, foldersQuery];
        if (currentSection === 'series') return [seriesQuery];
        if (currentSection === 'personal') return [personalQuery];
        if (currentSection === 'clips') return [clipsQuery];
        if (isFullPlaylistListNeeded) return [playlistsQuery];
        return [];
    }, [
        isHome,
        currentSection,
        isFullPlaylistListNeeded,
        seriesQuery,
        moviesQuery,
        personalQuery,
        clipsQuery,
        foldersQuery,
        hiddenQuery,
        playlistsQuery,
    ]);
    const isLoadingPlaylists = activeListQueries.some((listQuery) => listQuery.isLoading);

    const gridSectionQuery =
        currentSection === 'series'
            ? seriesQuery
            : currentSection === 'personal'
              ? personalQuery
              : currentSection === 'clips'
                ? clipsQuery
                : null;
    const hasMoreSectionPlaylists = gridSectionQuery?.hasNextPage ?? false;
    const isFetchingMoreSectionPlaylists = gridSectionQuery?.isFetchingNextPage ?? false;

    const loadedPlaylists = useMemo(
        () => [
            ...seriesPlaylists,
            ...moviePlaylists,
            ...personalPlaylists,
            ...clipPlaylists,
            ...folderPlaylists,
            ...playlists,
        ],
        [
            seriesPlaylists,
            moviePlaylists,
            personalPlaylists,
            clipPlaylists,
            folderPlaylists,
            playlists,
        ]
    );
    const loadedPlaylistForSelection = useMemo(() => {
        if (!playlistSlug) return null;
        const playlistById = playlistIdFromSearch
            ? loadedPlaylists.find((playlist) => playlist.id === playlistIdFromSearch)
            : undefined;
        return (
            playlistById ??
            loadedPlaylists.find(
                (playlist) => (slugify(playlist.name) || String(playlist.id)) === playlistSlug
            ) ??
            null
        );
    }, [loadedPlaylists, playlistIdFromSearch, playlistSlug]);

    const shouldFetchPlaylistHeaderById =
        Boolean(playlistSlug) &&
        playlistIdFromSearch !== null &&
        loadedPlaylistForSelection === null &&
        !isLoadingPlaylists;
    const playlistHeaderQuery = useQuery({
        queryKey: videoQueryKeys.playlistDetail(playlistIdFromSearch ?? undefined),
        enabled: shouldFetchPlaylistHeaderById,
        queryFn: () => getVideoPlaylistById(playlistIdFromSearch ?? 0),
    });
    const selectedPlaylistSummary = loadedPlaylistForSelection ?? playlistHeaderQuery.data ?? null;

    const selectedPlaylistQuery = useInfiniteQuery({
        queryKey: videoQueryKeys.playlistItems(selectedPlaylistSummary?.id),
        enabled: Boolean(selectedPlaylistSummary?.id),
        queryFn: ({ pageParam = 1 }) =>
            getVideoPlaylistItemsPage(
                selectedPlaylistSummary?.id ?? 0,
                pageParam,
                VIDEO_PLAYLIST_ITEMS_PAGE_SIZE
            ),
        initialPageParam: 1,
        getNextPageParam: (lastPage) =>
            lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined,
    });
    const {
        data: selectedPlaylistItemsData,
        isLoading: isLoadingSelectedPlaylist,
        isFetchingNextPage: isFetchingMoreSelectedPlaylistItems,
        hasNextPage: hasMoreSelectedPlaylistItems = false,
        fetchNextPage: fetchNextSelectedPlaylistItems,
    } = selectedPlaylistQuery;

    const selectedPlaylistDetail = useMemo<VideoPlaylistDto | null>(() => {
        if (!selectedPlaylistSummary || !selectedPlaylistItemsData) return null;
        return {
            ...selectedPlaylistSummary,
            items: selectedPlaylistItemsData.pages.flatMap((page) => page.items),
        };
    }, [selectedPlaylistSummary, selectedPlaylistItemsData]);

    const recentCatalogItems = useMemo(
        () => homeCatalog?.sections.find((section) => section.key === 'recent')?.items ?? [],
        [homeCatalog?.sections]
    );

    const allVideos = useMemo(
        () => videoLibraryData?.pages.flatMap((page) => page.items) ?? [],
        [videoLibraryData]
    );

    const filteredVideos = useMemo(() => {
        const normalizedSearch = videoSearch.trim().toLowerCase();
        if (!normalizedSearch) {
            return allVideos;
        }

        return allVideos.filter((video) =>
            [video.name, video.parent_path, video.format].some((value) =>
                value.toLowerCase().includes(normalizedSearch)
            )
        );
    }, [allVideos, videoSearch]);

    const invalidatePlaylistQueries = async () => {
        await Promise.all([
            queryClient.invalidateQueries({ queryKey: videoQueryKeys.playlists }),
            queryClient.invalidateQueries({ queryKey: ['video', 'playlist-detail'] }),
            queryClient.invalidateQueries({ queryKey: ['video', 'playlist-items'] }),
            queryClient.invalidateQueries({
                queryKey: ['video', 'playlists-of-video'],
            }),
        ]);
    };

    const invalidateAllVideoQueries = async () => {
        await Promise.all([
            invalidatePlaylistQueries(),
            queryClient.invalidateQueries({ queryKey: videoQueryKeys.homeCatalog }),
            queryClient.invalidateQueries({ queryKey: videoQueryKeys.continueWatching }),
        ]);
    };

    const addToPlaylistMutation = useMutation({
        mutationFn: async ({ playlistId, videoId }: { playlistId: number; videoId: number }) =>
            addVideoToPlaylist(playlistId, videoId),
        onSuccess: async () => {
            await invalidateAllVideoQueries();
            setFeedback({
                open: true,
                message: t('VIDEO_ADD_SUCCESS'),
                severity: 'success',
            });
        },
        onError: (error) => {
            setFeedback({
                open: true,
                message: extractBackendErrorMessage(error) ?? t('VIDEO_ADD_ERROR'),
                severity: 'error',
            });
        },
    });

    const createPlaylistMutation = useMutation({
        mutationFn: ({ name, videoId }: { name: string; videoId?: number }) =>
            createVideoPlaylist(name, videoId),
        onSuccess: async (_createdPlaylist, { videoId }) => {
            await invalidateAllVideoQueries();
            setFeedback({
                open: true,
                message: t(
                    videoId === undefined
                        ? 'VIDEO_PLAYLIST_CREATED'
                        : 'VIDEO_PLAYLIST_CREATED_ADDED'
                ),
                severity: 'success',
            });
        },
        onError: (error) => {
            setFeedback({
                open: true,
                message: extractBackendErrorMessage(error) ?? t('VIDEO_PLAYLIST_CREATE_ERROR'),
                severity: 'error',
            });
        },
    });

    const clearSelectedPlaylist = () => {
        if (currentSection === 'home') {
            navigate('/videos');
            return;
        }
        navigate(`/videos/${currentSection}`);
    };

    const deletePlaylistMutation = useMutation({
        mutationFn: (playlistId: number) => deleteVideoPlaylist(playlistId),
        onSuccess: async () => {
            clearSelectedPlaylist();
            await invalidateAllVideoQueries();
            setFeedback({ open: true, message: t('VIDEO_PLAYLIST_DELETED'), severity: 'success' });
        },
        onError: (error) => {
            setFeedback({
                open: true,
                message: extractBackendErrorMessage(error) ?? t('VIDEO_PLAYLIST_DELETE_ERROR'),
                severity: 'error',
            });
        },
    });

    const hidePlaylistFeedbackAction = (playlist: VideoPlaylistDto): VideoFeedbackAction => ({
        label: t('VIDEO_UNDO'),
        onClick: () => changeVisibilityMutation.mutate({ playlist, hidden: false }),
    });

    const changeVisibilityMutation = useMutation({
        mutationFn: ({ playlist, hidden }: { playlist: VideoPlaylistDto; hidden: boolean }) =>
            setVideoPlaylistHidden(playlist.id, hidden),
        onSuccess: async (_unused, { playlist, hidden }) => {
            if (hidden && selectedPlaylistSummary?.id === playlist.id) {
                clearSelectedPlaylist();
            }
            await invalidateAllVideoQueries();
            setFeedback(
                hidden
                    ? {
                          open: true,
                          message: t('VIDEO_PLAYLIST_HIDDEN'),
                          severity: 'success',
                          action: hidePlaylistFeedbackAction(playlist),
                      }
                    : { open: true, message: t('VIDEO_PLAYLIST_SHOWN'), severity: 'success' }
            );
        },
        onError: (error, { hidden }) => {
            setFeedback({
                open: true,
                message:
                    extractBackendErrorMessage(error) ??
                    t(hidden ? 'VIDEO_PLAYLIST_HIDE_ERROR' : 'VIDEO_PLAYLIST_SHOW_ERROR'),
                severity: 'error',
            });
        },
    });

    const renameMutation = useMutation({
        mutationFn: async (name: string) => {
            if (!selectedPlaylistSummary) return;
            return updateVideoPlaylistName(selectedPlaylistSummary.id, name);
        },
        onSuccess: () => invalidatePlaylistQueries(),
    });

    const removeFromPlaylistMutation = useMutation({
        mutationFn: async (videoId: number) => {
            if (!selectedPlaylistSummary) return;
            return removeVideoFromPlaylist(selectedPlaylistSummary.id, videoId);
        },
        onSuccess: () => invalidateAllVideoQueries(),
    });

    const reorderMutation = useMutation({
        mutationFn: async (items: { video_id: number; order_index: number }[]) => {
            if (!selectedPlaylistSummary) return;
            return reorderVideoPlaylist(selectedPlaylistSummary.id, items);
        },
        onSuccess: () => invalidatePlaylistQueries(),
    });

    const setWatchedMutation = useMutation({
        mutationFn: ({ videoId, watched }: { videoId: number; watched: boolean }) =>
            setVideoWatched(videoId, watched),
        onMutate: async ({ videoId, watched }) => {
            await Promise.all([
                queryClient.cancelQueries({ queryKey: videoQueryKeys.continueWatching }),
                queryClient.cancelQueries({ queryKey: ['video', 'playlist-items'] }),
            ]);
            const previousContinueWatching = queryClient.getQueryData<VideoContinueItemDto[]>(
                videoQueryKeys.continueWatching
            );
            const previousPlaylistItems = queryClient.getQueriesData<PlaylistItemsPages>({
                queryKey: ['video', 'playlist-items'],
            });

            queryClient.setQueryData<VideoContinueItemDto[]>(
                videoQueryKeys.continueWatching,
                (continueItems) =>
                    watched
                        ? continueItems?.filter((item) => item.video.id !== videoId)
                        : continueItems
            );
            queryClient.setQueriesData<PlaylistItemsPages>(
                { queryKey: ['video', 'playlist-items'] },
                (playlistItems) =>
                    playlistItems && {
                        ...playlistItems,
                        pages: playlistItems.pages.map((page) => ({
                            ...page,
                            items: page.items.map((item) =>
                                item.video.id === videoId
                                    ? {
                                          ...item,
                                          status: watched ? 'completed' : 'not_started',
                                          progress_pct: watched ? 100 : 0,
                                      }
                                    : item
                            ),
                        })),
                    }
            );

            return { previousContinueWatching, previousPlaylistItems };
        },
        onError: (_error, _variables, rollbackContext) => {
            queryClient.setQueryData(
                videoQueryKeys.continueWatching,
                rollbackContext?.previousContinueWatching
            );
            rollbackContext?.previousPlaylistItems.forEach(([queryKey, playlistItems]) =>
                queryClient.setQueryData(queryKey, playlistItems)
            );
            setFeedback({
                open: true,
                message: t('VIDEO_WATCHED_ERROR'),
                severity: 'error',
            });
        },
        onSettled: () => invalidateAllVideoQueries(),
    });

    const playlistsFailure =
        activeListQueries
            .map((listQuery) =>
                toVideoQueryFailure({
                    isError: listQuery.isError,
                    error: listQuery.error,
                    hasData: listQuery.data !== undefined,
                    refetch: listQuery.refetch,
                })
            )
            .find((failure) => failure !== null) ?? null;
    const homeCatalogFailure = toVideoQueryFailure({
        isError: homeCatalogQuery.isError,
        error: homeCatalogQuery.error,
        hasData: homeCatalogQuery.data !== undefined,
        refetch: homeCatalogQuery.refetch,
    });
    const videosFailure = toVideoQueryFailure({
        isError: videoLibraryQuery.isError,
        error: videoLibraryQuery.error,
        hasData: videoLibraryQuery.data !== undefined,
        refetch: videoLibraryQuery.refetch,
    });
    const continueWatchingFailure = toVideoQueryFailure({
        isError: continueWatchingQuery.isError,
        error: continueWatchingQuery.error,
        hasData: continueWatchingQuery.data !== undefined,
        refetch: continueWatchingQuery.refetch,
    });
    const selectedPlaylistFailure = toVideoQueryFailure({
        isError: selectedPlaylistQuery.isError,
        error: selectedPlaylistQuery.error,
        hasData: selectedPlaylistQuery.data !== undefined,
        refetch: selectedPlaylistQuery.refetch,
    });

    const getCurrentRoute = () => `${location.pathname}${location.search}`;

    const resolvePlaylistSection = (playlist: VideoPlaylistDto): Exclude<VideoSection, 'home'> =>
        currentSection !== 'home' ? currentSection : getVideoSectionForPlaylist(playlist);

    const buildVideoPlayerUrl = (videoId: number, playlistId?: number | null) => {
        const params = new URLSearchParams();
        if (playlistId) {
            params.set('playlist', String(playlistId));
        }
        const from = getCurrentRoute();
        if (from) {
            params.set('from', from);
        }
        const qs = params.toString();
        return `/video/${videoId}${qs ? `?${qs}` : ''}`;
    };

    const playVideo = (videoId: number, playlistId?: number | null) => {
        if (!videoId) return;
        const from = getCurrentRoute();
        navigate(buildVideoPlayerUrl(videoId, playlistId), {
            state: { from, playlistId: playlistId ?? null },
        });
    };

    const openPlaylistVideo = (videoId: number) => {
        if (!selectedPlaylistSummary) return;
        const from = getCurrentRoute();
        navigate(buildVideoPlayerUrl(videoId, selectedPlaylistSummary.id), {
            state: {
                from,
                playlistId: selectedPlaylistSummary.id,
            },
        });
    };

    const contextValue: VideoContentContextData = {
        currentSection,
        playlists,
        allVideos,
        filteredVideos,
        continueWatchingItems,
        seriesPlaylists,
        moviePlaylists,
        personalPlaylists,
        clipPlaylists,
        folderPlaylists,
        hiddenPlaylists,
        recentCatalogItems,
        selectedPlaylistSummary,
        selectedPlaylistDetail,
        isLoadingPlaylists,
        isLoadingVideos,
        isLoadingSelectedPlaylist,
        isFetchingMoreSelectedPlaylistItems,
        hasMoreSelectedPlaylistItems,
        isLoadingHomeCatalog,
        isLoadingContinueWatching,
        playlistsFailure,
        videosFailure,
        selectedPlaylistFailure,
        homeCatalogFailure,
        continueWatchingFailure,
        isFetchingMoreVideos,
        hasMoreVideos,
        isFetchingMoreSectionPlaylists,
        hasMoreSectionPlaylists,
        isFetchingMoreHiddenPlaylists: hiddenQuery.isFetchingNextPage,
        hasMoreHiddenPlaylists: hiddenQuery.hasNextPage ?? false,
        isChangingPlaylistVisibility: changeVisibilityMutation.isPending,
        isAddingToPlaylist: addToPlaylistMutation.isPending,
        isCreatingPlaylist: createPlaylistMutation.isPending,
        isDeletingPlaylist: deletePlaylistMutation.isPending,
        isRenamingPlaylist: renameMutation.isPending,
        isRemovingFromPlaylist: removeFromPlaylistMutation.isPending,
        isReorderingPlaylist: reorderMutation.isPending,
        videoSearch,
        librarySort,
        libraryFilters,
        feedback,
        setVideoSearch,
        setLibrarySort,
        setLibraryFilters,
        closeFeedback: () => setFeedback((prev) => ({ ...prev, open: false })),
        loadMoreVideos: () => {
            if (hasMoreVideos && !isFetchingMoreVideos) {
                void fetchNextPage();
            }
        },
        loadMoreSectionPlaylists: () => {
            if (gridSectionQuery?.hasNextPage && !gridSectionQuery.isFetchingNextPage) {
                void gridSectionQuery.fetchNextPage();
            }
        },
        loadMoreHiddenPlaylists: () => {
            if (hiddenQuery.hasNextPage && !hiddenQuery.isFetchingNextPage) {
                void hiddenQuery.fetchNextPage();
            }
        },
        loadMoreSelectedPlaylistItems: () => {
            if (hasMoreSelectedPlaylistItems && !isFetchingMoreSelectedPlaylistItems) {
                void fetchNextSelectedPlaylistItems();
            }
        },
        selectPlaylist: (playlist) =>
            navigate(
                getVideoDetailRoute(
                    resolvePlaylistSection(playlist),
                    slugify(playlist.name) || String(playlist.id),
                    playlist.id
                )
            ),
        clearSelectedPlaylist,
        playVideo,
        openPlaylistVideo,
        addVideoFromLibrary: (videoId, playlistId) =>
            addToPlaylistMutation.mutate({ playlistId, videoId }),
        createPlaylist: (name) => createPlaylistMutation.mutate({ name }),
        createPlaylistWithVideo: (name, videoId) =>
            createPlaylistMutation.mutate({ name, videoId }),
        deleteSelectedPlaylist: () => {
            if (!selectedPlaylistSummary) return;
            deletePlaylistMutation.mutate(selectedPlaylistSummary.id);
        },
        hidePlaylist: (playlist) => changeVisibilityMutation.mutate({ playlist, hidden: true }),
        hideSelectedPlaylist: () => {
            if (!selectedPlaylistSummary) return;
            changeVisibilityMutation.mutate({ playlist: selectedPlaylistSummary, hidden: true });
        },
        showPlaylist: (playlist) => changeVisibilityMutation.mutate({ playlist, hidden: false }),
        renameSelectedPlaylist: (name) => {
            if (!name.trim()) return;
            renameMutation.mutate(name);
        },
        removeVideoFromSelectedPlaylist: (videoId) => removeFromPlaylistMutation.mutate(videoId),
        setVideoWatched: (videoId, watched) => setWatchedMutation.mutate({ videoId, watched }),
        moveSelectedPlaylistItem: (index, direction) => {
            if (!selectedPlaylistDetail) return;
            const orderedItems = [...selectedPlaylistDetail.items].sort(
                (a, b) => a.order_index - b.order_index || a.id - b.id
            );
            const current = orderedItems[index];
            const neighbor = orderedItems[index + direction];
            if (!current || !neighbor) return;
            reorderMutation.mutate([
                { video_id: current.video.id, order_index: neighbor.order_index },
                { video_id: neighbor.video.id, order_index: current.order_index },
            ]);
        },
    };

    return (
        <VideoContentContext.Provider value={contextValue}>{children}</VideoContentContext.Provider>
    );
}

export function useVideoContentProvider() {
    const context = useContext(VideoContentContext);
    if (!context) {
        throw new Error('useVideoContentProvider must be used within VideoContentProvider');
    }
    return context;
}
