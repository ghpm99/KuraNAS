/* eslint-disable react-refresh/only-export-components */
import {
	addVideoToPlaylist,
	getVideoHomeCatalog,
	getVideoLibraryFiles,
	getVideoPlaylistMemberships,
	getVideoPlaybackState,
	getVideoPlaylistById,
	getVideoPlaylists,
	type VideoCatalogItemDto,
	reorderVideoPlaylist,
	removeVideoFromPlaylist,
	updateVideoPlaylistName,
	type VideoFileDto,
	type VideoPlaylistDto,
} from '@/service/videoPlayback';
import { type VideoSection } from '@/app/routes';
import {
	getVideoDetailRoute,
	getVideoDetailSlugFromPath,
	getVideoSectionForPlaylist,
	getVideoSectionFromPath,
} from '@/components/videos/navigation';
import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { createContext, useContext, useMemo, useState, type ReactNode } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import useI18n from '@/components/i18n/provider/i18nContext';

const VIDEO_LIBRARY_PAGE_SIZE = 60;
const VIDEO_HOME_CATALOG_LIMIT = 12;

type FeedbackState = { open: boolean; message: string; severity: 'success' | 'error' };

export interface VideoContentContextData {
	currentSection: VideoSection;
	playlists: VideoPlaylistDto[];
	allVideos: VideoFileDto[];
	filteredVideos: VideoFileDto[];
	continuePlaylists: VideoPlaylistDto[];
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
	isLoadingHomeCatalog: boolean;
	isFetchingMoreVideos: boolean;
	hasMoreVideos: boolean;
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
	selectPlaylist: (playlist: VideoPlaylistDto) => void;
	clearSelectedPlaylist: () => void;
	playVideo: (videoId: number, playlistId?: number | null) => void;
	openPlaylistVideo: (videoId: number) => void;
	addVideoFromLibrary: (videoId: number) => void;
	renameSelectedPlaylist: (name: string) => void;
	removeVideoFromSelectedPlaylist: (videoId: number) => void;
	moveSelectedPlaylistItem: (index: number, direction: -1 | 1) => void;
}

const VideoContentContext = createContext<VideoContentContextData | undefined>(undefined);

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
	const [videoSearch, setVideoSearch] = useState('');
	const [selectedPlaylistPerVideo, setSelectedPlaylistPerVideo] = useState<Record<number, number>>({});
	const [feedback, setFeedback] = useState<FeedbackState>({ open: false, message: '', severity: 'success' });
	const currentSection = getVideoSectionFromPath(location.pathname);

	const { data: playlists = [], isLoading: isLoadingPlaylists } = useQuery({
		queryKey: ['video', 'playlists'],
		queryFn: () => getVideoPlaylists(false),
	});
	const { data: homeCatalog, isLoading: isLoadingHomeCatalog } = useQuery({
		queryKey: ['video', 'home-catalog'],
		queryFn: () => getVideoHomeCatalog(VIDEO_HOME_CATALOG_LIMIT),
	});
	const {
		data: videoLibraryData,
		isLoading: isLoadingVideos,
		isFetchingNextPage: isFetchingMoreVideos,
		hasNextPage: hasMoreVideos = false,
		fetchNextPage,
	} = useInfiniteQuery({
		queryKey: ['video', 'library-files', videoSearch],
		queryFn: ({ pageParam = 1 }) => getVideoLibraryFiles(pageParam, VIDEO_LIBRARY_PAGE_SIZE, videoSearch),
		initialPageParam: 1,
		getNextPageParam: (lastPage) => (lastPage.pagination.has_next ? lastPage.pagination.page + 1 : undefined),
	});

	const hasContinuePlaylists = useMemo(
		() => playlists.some((playlist) => Boolean(playlist.last_played_at)),
		[playlists],
	);

	const { data: playbackState } = useQuery({
		queryKey: ['video', 'playback-state'],
		queryFn: getVideoPlaybackState,
		retry: false,
		enabled: hasContinuePlaylists,
	});

	const playlistSlug = getVideoDetailSlugFromPath(location.pathname);
	const selectedPlaylistSummary = useMemo(() => {
		if (!playlistSlug) return null;
		return playlists.find((playlist) => slugify(playlist.name) === playlistSlug) ?? null;
	}, [playlistSlug, playlists]);

	const { data: selectedPlaylistDetailData, isLoading: isLoadingSelectedPlaylist } = useQuery({
		queryKey: ['video', 'playlist-detail', selectedPlaylistSummary?.id],
		enabled: Boolean(selectedPlaylistSummary?.id),
		queryFn: () => getVideoPlaylistById(selectedPlaylistSummary?.id ?? 0),
	});

	const selectedPlaylistDetail = selectedPlaylistDetailData ?? null;

	const continuePlaylists = useMemo(() => {
		const sorted = [...playlists]
			.filter((playlist) => Boolean(playlist.last_played_at))
			.sort((a, b) => {
				const aTime = a.last_played_at ? new Date(a.last_played_at).getTime() : 0;
				const bTime = b.last_played_at ? new Date(b.last_played_at).getTime() : 0;
				return bTime - aTime;
			});

		const playbackPlaylistId = playbackState?.playback_state.playlist_id;
		const playbackVideoId = playbackState?.playback_state.video_id;
		if (!playbackPlaylistId) return sorted;

		return sorted.map((playlist) => {
			if (playlist.id !== playbackPlaylistId) return playlist;
			return { ...playlist, cover_video_id: playbackVideoId ?? playlist.cover_video_id };
		});
	}, [playlists, playbackState]);

	const seriesPlaylists = useMemo(
		() => playlists.filter((playlist) => playlist.classification === 'series' || playlist.classification === 'anime'),
		[playlists],
	);

	const moviePlaylists = useMemo(
		() => playlists.filter((playlist) => playlist.classification === 'movie'),
		[playlists],
	);

	const personalPlaylists = useMemo(
		() => playlists.filter((playlist) => playlist.classification === 'personal'),
		[playlists],
	);

	const clipPlaylists = useMemo(
		() => playlists.filter((playlist) => playlist.classification === 'clip' || playlist.classification === 'program'),
		[playlists],
	);

	const folderPlaylists = useMemo(() => playlists.filter((playlist) => playlist.type === 'folder'), [playlists]);

	const recentCatalogItems = useMemo(
		() => homeCatalog?.sections.find((section) => section.key === 'recent')?.items ?? [],
		[homeCatalog?.sections],
	);

	const allVideos = useMemo(
		() => videoLibraryData?.pages.flatMap((page) => page.items) ?? [],
		[videoLibraryData],
	);

	const filteredVideos = useMemo(() => {
		const normalizedSearch = videoSearch.trim().toLowerCase();
		if (!normalizedSearch) {
			return allVideos;
		}

		return allVideos.filter((video) =>
			[video.name, video.parent_path, video.format].some((value) => value.toLowerCase().includes(normalizedSearch)),
		);
	}, [allVideos, videoSearch]);

	const { data: playlistMemberships = [] } = useQuery({
		queryKey: ['video', 'playlist-membership', playlists.map((playlist) => playlist.id).join(',')],
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
			queryClient.invalidateQueries({ queryKey: ['video', 'playlists'] }),
			queryClient.invalidateQueries({ queryKey: ['video', 'playlist-detail'] }),
			queryClient.invalidateQueries({ queryKey: ['video', 'playlist-membership'] }),
		]);
	};

	const invalidateAllVideoQueries = async () => {
		await Promise.all([
			invalidatePlaylistQueries(),
			queryClient.invalidateQueries({ queryKey: ['video', 'home-catalog'] }),
		]);
	};

	const addToPlaylistMutation = useMutation({
		mutationFn: async ({ playlistId, videoId }: { playlistId: number; videoId: number }) =>
			addVideoToPlaylist(playlistId, videoId),
		onSuccess: async () => {
			await invalidateAllVideoQueries();
			setFeedback({ open: true, message: t('VIDEO_ADD_SUCCESS'), severity: 'success' });
		},
		onError: () => {
			setFeedback({ open: true, message: t('VIDEO_ADD_ERROR'), severity: 'error' });
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

	const getCurrentRoute = () => `${location.pathname}${location.search}`;

	const resolvePlaylistSection = (playlist: VideoPlaylistDto): Exclude<VideoSection, 'home'> =>
		currentSection !== 'home' ? currentSection : getVideoSectionForPlaylist(playlist);

	const playVideo = (videoId: number, playlistId?: number | null) => {
		if (!videoId) return;
		const from = getCurrentRoute();
		navigate(`/video/${videoId}`, { state: { from, playlistId: playlistId ?? null } });
	};

	const openPlaylistVideo = (videoId: number) => {
		if (!selectedPlaylistSummary) return;
		navigate(`/video/${videoId}`, {
			state: {
				from: getCurrentRoute(),
				playlistId: selectedPlaylistSummary.id,
			},
		});
	};

	const contextValue: VideoContentContextData = {
		currentSection,
		playlists,
		allVideos,
		filteredVideos,
		continuePlaylists,
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
		isLoadingHomeCatalog,
		isFetchingMoreVideos,
		hasMoreVideos,
		isAddingToPlaylist: addToPlaylistMutation.isPending,
		isRenamingPlaylist: renameMutation.isPending,
		isRemovingFromPlaylist: removeFromPlaylistMutation.isPending,
		isReorderingPlaylist: reorderMutation.isPending,
		videoSearch,
		selectedPlaylistPerVideo,
		feedback,
		setVideoSearch,
		setSelectedPlaylistForVideo: (videoId, playlistId) => {
			setSelectedPlaylistPerVideo((prev) => ({ ...prev, [videoId]: playlistId }));
		},
		closeFeedback: () => setFeedback((prev) => ({ ...prev, open: false })),
		loadMoreVideos: () => {
			if (hasMoreVideos && !isFetchingMoreVideos) {
				void fetchNextPage();
			}
		},
		selectPlaylist: (playlist) => navigate(getVideoDetailRoute(resolvePlaylistSection(playlist), slugify(playlist.name))),
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
		moveSelectedPlaylistItem: (index, direction) => {
			if (!selectedPlaylistDetail) return;
			const orderedItems = [...selectedPlaylistDetail.items].sort((a, b) => a.order_index - b.order_index || a.id - b.id);
			const target = index + direction;
			if (target < 0 || target >= orderedItems.length) return;
			const swapped = [...orderedItems];
			const current = swapped[index];
			const next = swapped[target];
			if (!current || !next) return;
			swapped[index] = next;
			swapped[target] = current;
			reorderMutation.mutate(
				swapped.map((item, idx) => ({
					video_id: item.video.id,
					order_index: idx,
				})),
			);
		},
	};

	return <VideoContentContext.Provider value={contextValue}>{children}</VideoContentContext.Provider>;
}

export function useVideoContentProvider() {
	const context = useContext(VideoContentContext);
	if (!context) {
		throw new Error('useVideoContentProvider must be used within VideoContentProvider');
	}
	return context;
}
