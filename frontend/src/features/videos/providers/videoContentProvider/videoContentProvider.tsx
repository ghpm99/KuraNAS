/* eslint-disable react-refresh/only-export-components */
import {
    addVideoToPlaylist,
    getVideoContinueWatching,
    getVideoHomeCatalog,
    getVideoLibraryFiles,
    getVideoPlaylistMemberships,
    getVideoPlaylistById,
    getVideoPlaylistItemsPage,
    getVideoPlaylists,
    type VideoCatalogItemDto,
    type VideoContinueItemDto,
    reorderVideoPlaylist,
    removeVideoFromPlaylist,
    setVideoWatched,
    updateVideoPlaylistName,
    type VideoFileDto,
    type VideoPlaylistDto,
    type VideoPlaylistItemDto,
} from '@/service/videoPlayback';
import { useVideoSectionPlaylists, videoQueryKeys } from './useVideoQueries';
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
import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import useI18n from '@/components/i18n/provider/i18nContext';

const VIDEO_LIBRARY_PAGE_SIZE = 60;
const VIDEO_PLAYLIST_ITEMS_PAGE_SIZE = 50;
const VIDEO_SECTION_GRID_PAGE_SIZE = 24;
const VIDEO_HOME_RAIL_PAGE_SIZE = 4;
const VIDEO_HOME_CATALOG_LIMIT = 12;
const VIDEO_CONTINUE_WATCHING_LIMIT = 24;

type PlaylistItemsPages = InfiniteData<Pagination<VideoPlaylistItemDto>>;

type FeedbackState = {
    open: boolean;
    message: string;
    severity: 'success' | 'error';
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
    recentCatalogItems: VideoCatalogItemDto[];
    playlistMembershipMap: Record<number, Set<number>>;
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
    isAddingToPlaylist: boolean;
    isRenamingPlaylist: boolean;
    isRemovingFromPlaylist: boolean;
    isReorderingPlaylist: boolean;
    videoSearch: string;
    selectedPlaylistPerVideo: Record<number, number>;
    feedback: FeedbackState;
    setVideoSearch: (value: string) => void;
    setSelectedPlaylistForVideo: (videoId: number, playlistId: number) => void;
    closeFeedback: () => void;
    loadMoreVideos: () => void;
    loadMoreSectionPlaylists: () => void;
    loadMoreSelectedPlaylistItems: () => void;
    selectPlaylist: (playlist: VideoPlaylistDto) => void;
    clearSelectedPlaylist: () => void;
    playVideo: (videoId: number, playlistId?: number | null) => void;
    openPlaylistVideo: (videoId: number) => void;
    addVideoFromLibrary: (videoId: number) => void;
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
    const searchTextFromUrl = new URLSearchParams(location.search).get('q') ?? '';
    const [videoSearch, setVideoSearch] = useState(searchTextFromUrl);
    useEffect(() => {
        setVideoSearch(searchTextFromUrl);
    }, [searchTextFromUrl]);
    const [selectedPlaylistPerVideo, setSelectedPlaylistPerVideo] = useState<
        Record<number, number>
    >({});
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
        queryKey: videoQueryKeys.libraryFiles(videoSearch),
        queryFn: ({ pageParam = 1 }) =>
            getVideoLibraryFiles(pageParam, VIDEO_LIBRARY_PAGE_SIZE, videoSearch),
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

    const activeListQueries = useMemo(() => {
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

    const { data: playlistMemberships = [] } = useQuery({
        queryKey: videoQueryKeys.playlistMembership(
            playlists.map((playlist) => playlist.id).join(',')
        ),
        enabled: playlists.length > 0,
        queryFn: () => getVideoPlaylistMemberships(false),
    });

    const playlistMembershipMap = useMemo<Record<number, Set<number>>>(() => {
        const membershipsByPlaylist: Record<number, Set<number>> = {};
        for (const membership of playlistMemberships) {
            if (!membershipsByPlaylist[membership.playlist_id]) {
                membershipsByPlaylist[membership.playlist_id] = new Set<number>();
            }
            membershipsByPlaylist[membership.playlist_id]?.add(membership.video_id);
        }

        return membershipsByPlaylist;
    }, [playlistMemberships]);

    const invalidatePlaylistQueries = async () => {
        await Promise.all([
            queryClient.invalidateQueries({ queryKey: videoQueryKeys.playlists }),
            queryClient.invalidateQueries({ queryKey: ['video', 'playlist-detail'] }),
            queryClient.invalidateQueries({ queryKey: ['video', 'playlist-items'] }),
            queryClient.invalidateQueries({
                queryKey: ['video', 'playlist-membership'],
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
        onError: () => {
            setFeedback({
                open: true,
                message: t('VIDEO_ADD_ERROR'),
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
        recentCatalogItems,
        playlistMembershipMap,
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
        isAddingToPlaylist: addToPlaylistMutation.isPending,
        isRenamingPlaylist: renameMutation.isPending,
        isRemovingFromPlaylist: removeFromPlaylistMutation.isPending,
        isReorderingPlaylist: reorderMutation.isPending,
        videoSearch,
        selectedPlaylistPerVideo,
        feedback,
        setVideoSearch,
        setSelectedPlaylistForVideo: (videoId, playlistId) => {
            setSelectedPlaylistPerVideo((prev) => ({
                ...prev,
                [videoId]: playlistId,
            }));
        },
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
        clearSelectedPlaylist: () => {
            if (currentSection === 'home') {
                navigate('/videos');
                return;
            }
            navigate(`/videos/${currentSection}`);
        },
        playVideo,
        openPlaylistVideo,
        addVideoFromLibrary: (videoId) => {
            const playlistId = selectedPlaylistPerVideo[videoId] ?? playlists[0]?.id;
            if (!playlistId) return;
            addToPlaylistMutation.mutate({ playlistId, videoId });
        },
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
