/* eslint-disable react-refresh/only-export-components */
import {
	addVideoToPlaylist,
	getAllVideoFiles,
	getVideoHomeCatalog,
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
import { getVideoSectionFromPath } from '@/components/videos/navigation';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { createContext, useContext, useMemo, useState, type ReactNode } from 'react';
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom';
import useI18n from '@/components/i18n/provider/i18nContext';

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
	const [searchParams, setSearchParams] = useSearchParams();
	const navigate = useNavigate();
	const location = useLocation();
	const queryClient = useQueryClient();
	const [videoSearch, setVideoSearch] = useState('');
	const [selectedPlaylistPerVideo, setSelectedPlaylistPerVideo] = useState<Record<number, number>>({});
	const [feedback, setFeedback] = useState<FeedbackState>({ open: false, message: '', severity: 'success' });
	const currentSection = getVideoSectionFromPath(location.pathname);

	const { data: playlists = [], isLoading: isLoadingPlaylists } = useQuery({
		queryKey: ['video-playlists'],
		queryFn: () => getVideoPlaylists(false),
	});
	const { data: allVideos = [], isLoading: isLoadingVideos } = useQuery({
		queryKey: ['video-all-files'],
		queryFn: () => getAllVideoFiles(2000),
	});
	const { data: homeCatalog, isLoading: isLoadingHomeCatalog } = useQuery({
		queryKey: ['video-home-catalog'],
		queryFn: () => getVideoHomeCatalog(12),
	});
	const { data: playbackState } = useQuery({
		queryKey: ['video-playback-state'],
		queryFn: getVideoPlaybackState,
		retry: false,
	});

	const playlistSlug = searchParams.get('playlist') || '';
	const selectedPlaylistSummary = useMemo(() => {
		if (!playlistSlug) return null;
		return playlists.find((playlist) => slugify(playlist.name) === playlistSlug) ?? null;
	}, [playlistSlug, playlists]);

	const { data: selectedPlaylistDetailData, isLoading: isLoadingSelectedPlaylist } = useQuery({
		queryKey: ['video-playlist', selectedPlaylistSummary?.id],
		enabled: Boolean(selectedPlaylistSummary?.id),
		queryFn: () => getVideoPlaylistById(selectedPlaylistSummary?.id ?? 0),
	});

	const selectedPlaylistDetail = selectedPlaylistDetailData ?? null;

	const continuePlaylists = useMemo(
		() =>
			[...playlists]
				.filter((playlist) => Boolean(playlist.last_played_at))
				.sort((a, b) => {
					const aTime = a.last_played_at ? new Date(a.last_played_at).getTime() : 0;
					const bTime = b.last_played_at ? new Date(b.last_played_at).getTime() : 0;
					return bTime - aTime;
				}),
		[playlists],
	);

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

	const filteredVideos = useMemo(() => {
		if (!videoSearch.trim()) return allVideos;
		const query = videoSearch.toLowerCase();
		return allVideos.filter(
			(video) =>
				video.name.toLowerCase().includes(query) ||
				video.parent_path.toLowerCase().includes(query) ||
				video.format.toLowerCase().includes(query),
		);
	}, [allVideos, videoSearch]);

	const { data: playlistMembershipMap = {} } = useQuery({
		queryKey: ['video-playlist-membership', playlists.map((playlist) => playlist.id).join(',')],
		enabled: playlists.length > 0,
		queryFn: async () => {
			const entries = await Promise.all(
				playlists.map(async (playlist) => {
					const detail = await getVideoPlaylistById(playlist.id);
					return [playlist.id, new Set(detail.items.map((item) => item.video.id))] as const;
				}),
			);
			return Object.fromEntries(entries) as Record<number, Set<number>>;
		},
	});

	const refreshVideoQueries = async () => {
		await Promise.all([
			queryClient.invalidateQueries({ queryKey: ['video-playlists'] }),
			queryClient.invalidateQueries({ queryKey: ['video-playlist'] }),
			queryClient.invalidateQueries({ queryKey: ['video-playlist-membership'] }),
			queryClient.invalidateQueries({ queryKey: ['video-home-catalog'] }),
		]);
	};

	const addToPlaylistMutation = useMutation({
		mutationFn: async ({ playlistId, videoId }: { playlistId: number; videoId: number }) =>
			addVideoToPlaylist(playlistId, videoId),
		onSuccess: async () => {
			await refreshVideoQueries();
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
		onSuccess: refreshVideoQueries,
	});

	const removeFromPlaylistMutation = useMutation({
		mutationFn: async (videoId: number) => {
			if (!selectedPlaylistSummary) return;
			return removeVideoFromPlaylist(selectedPlaylistSummary.id, videoId);
		},
		onSuccess: refreshVideoQueries,
	});

	const reorderMutation = useMutation({
		mutationFn: async (items: { video_id: number; order_index: number }[]) => {
			if (!selectedPlaylistSummary) return;
			return reorderVideoPlaylist(selectedPlaylistSummary.id, items);
		},
		onSuccess: refreshVideoQueries,
	});

	const getCurrentRoute = () => {
		const currentSearch = searchParams.toString();
		return `${location.pathname}${currentSearch ? `?${currentSearch}` : ''}`;
	};

	const playVideo = (videoId: number, playlistId?: number | null) => {
		if (!videoId) return;
		const from = getCurrentRoute();
		navigate(`/video/${videoId}`, { state: { from, playlistId: playlistId ?? null } });
	};

	const openPlaylistVideo = (videoId: number) => {
		if (!selectedPlaylistSummary) return;
		const slug = slugify(selectedPlaylistSummary.name);
		setSearchParams({ playlist: slug, video: String(videoId) });
		navigate(`/video/${videoId}`, {
			state: {
				from: `${location.pathname}?playlist=${encodeURIComponent(slug)}&video=${videoId}`,
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
		selectPlaylist: (playlist) => setSearchParams({ playlist: slugify(playlist.name) }),
		clearSelectedPlaylist: () => setSearchParams({}),
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

	const playbackPlaylistId = playbackState?.playback_state.playlist_id;
	const playbackVideoId = playbackState?.playback_state.video_id;
	contextValue.continuePlaylists = continuePlaylists.map((playlist) => {
		if (playlist.id !== playbackPlaylistId) {
			return playlist;
		}
		return { ...playlist, cover_video_id: playbackVideoId ?? playlist.cover_video_id };
	});

	return <VideoContentContext.Provider value={contextValue}>{children}</VideoContentContext.Provider>;
}

export function useVideoContentProvider() {
	const context = useContext(VideoContentContext);
	if (!context) {
		throw new Error('useVideoContentProvider must be used within VideoContentProvider');
	}
	return context;
}
