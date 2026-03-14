import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';
import type { IMusicData } from '@/components/providers/musicProvider/musicProvider';
import { fetchAnalyticsOverview } from '@/service/analytics';
import { getPlayerState } from '@/service/playerState';
import { getNowPlayingPlaylist, getPlaylistTracks } from '@/service/playlist';
import { getVideoHomeCatalog, getVideoPlaybackState, type VideoCatalogItemDto, type VideoFileDto } from '@/service/videoPlayback';
import { useQuery } from '@tanstack/react-query';
import { useMemo, useState } from 'react';

const homeAnalyticsPeriod = '30d' as const;
const nowPlayingPageSize = 200;
const videoHomeLimit = 12;

const clampProgress = (value: number) => {
	if (!Number.isFinite(value)) {
		return 0;
	}

	return Math.max(0, Math.min(100, value));
};

const getProgressPercent = (currentTime: number, duration: number, fallback = 0) => {
	if (duration > 0) {
		return clampProgress((currentTime / duration) * 100);
	}

	return clampProgress(fallback);
};

export type HomeMusicResume = {
	track: IMusicData;
	progressSeconds: number;
	durationSeconds: number;
	progressPercent: number;
	queueCount: number;
	isPlaying: boolean;
};

export type HomeVideoResume = {
	video: VideoFileDto;
	progressSeconds: number;
	durationSeconds: number;
	progressPercent: number;
	playlistId: number | null;
};

const useHomeScreen = () => {
	const [searchQuery, setSearchQuery] = useState('');
	const {
		queue,
		currentTrack,
		currentTime,
		duration,
		isPlaying,
	} = useGlobalMusic();

	const analyticsQuery = useQuery({
		queryKey: ['home', 'analytics-overview', homeAnalyticsPeriod],
		queryFn: () => fetchAnalyticsOverview(homeAnalyticsPeriod),
	});

	const videoCatalogQuery = useQuery({
		queryKey: ['home', 'video-home-catalog'],
		queryFn: () => getVideoHomeCatalog(videoHomeLimit),
	});

	const videoPlaybackQuery = useQuery({
		queryKey: ['home', 'video-playback-state'],
		queryFn: () => getVideoPlaybackState(),
		retry: false,
	});

	const playerStateQuery = useQuery({
		queryKey: ['home', 'music-player-state'],
		queryFn: () => getPlayerState(),
		retry: false,
	});

	const nowPlayingQuery = useQuery({
		queryKey: ['home', 'music-now-playing'],
		queryFn: () => getNowPlayingPlaylist(),
		retry: false,
	});

	const nowPlayingTracksQuery = useQuery({
		queryKey: ['home', 'music-now-playing-tracks', nowPlayingQuery.data?.id],
		queryFn: () => getPlaylistTracks(nowPlayingQuery.data!.id, 1, nowPlayingPageSize),
		enabled: Boolean(nowPlayingQuery.data?.id),
		retry: false,
	});

	const recentFiles = analyticsQuery.data?.recent_files?.slice(0, 6) ?? [];

	const videoContinueItems = useMemo(() => {
		const continueSection = videoCatalogQuery.data?.sections.find((section) => section.key === 'continue');
		return continueSection?.items.slice(0, 4) ?? [];
	}, [videoCatalogQuery.data]);

	const videoResume = useMemo<HomeVideoResume | null>(() => {
		const session = videoPlaybackQuery.data;
		const videoId = session?.playback_state.video_id;
		if (!session || !videoId) {
			return null;
		}

		const activeItem = session.playlist.items.find((item) => item.video.id === videoId);
		if (!activeItem) {
			return null;
		}

		return {
			video: activeItem.video,
			progressSeconds: session.playback_state.current_time,
			durationSeconds: session.playback_state.duration,
			progressPercent: getProgressPercent(
				session.playback_state.current_time,
				session.playback_state.duration,
			),
			playlistId: session.playback_state.playlist_id,
		};
	}, [videoPlaybackQuery.data]);

	const fallbackMusicTrack = useMemo(() => {
		const currentFileId = playerStateQuery.data?.current_file_id;
		if (!currentFileId) {
			return null;
		}

		return nowPlayingTracksQuery.data?.items.find((item) => item.file.id === currentFileId)?.file ?? null;
	}, [nowPlayingTracksQuery.data, playerStateQuery.data?.current_file_id]);

	const musicResume = useMemo<HomeMusicResume | null>(() => {
		const activeTrack = currentTrack ?? fallbackMusicTrack;
		if (!activeTrack) {
			return null;
		}

		const progressSeconds = currentTrack ? currentTime : (playerStateQuery.data?.current_position ?? 0);
		const durationSeconds = currentTrack
			? Math.max(duration, activeTrack.metadata?.duration ?? 0)
			: (activeTrack.metadata?.duration ?? 0);
		const queueCount = queue.length || nowPlayingQuery.data?.track_count || nowPlayingTracksQuery.data?.items.length || 0;

		return {
			track: activeTrack,
			progressSeconds,
			durationSeconds,
			progressPercent: getProgressPercent(progressSeconds, durationSeconds),
			queueCount,
			isPlaying: currentTrack ? isPlaying : false,
		};
	}, [
		currentTime,
		currentTrack,
		duration,
		fallbackMusicTrack,
		isPlaying,
		nowPlayingQuery.data?.track_count,
		nowPlayingTracksQuery.data?.items.length,
		playerStateQuery.data?.current_position,
		queue.length,
	]);

	return {
		searchQuery,
		setSearchQuery,
		recentFiles,
		videoContinueItems,
		videoResume,
		musicResume,
		analytics: analyticsQuery.data,
		isAnalyticsLoading: analyticsQuery.isLoading,
		isVideoLoading: videoCatalogQuery.isLoading || videoPlaybackQuery.isLoading,
		isMusicLoading: playerStateQuery.isLoading || nowPlayingQuery.isLoading || nowPlayingTracksQuery.isLoading,
	};
};

export const homeScreenUtils = {
	getProgressPercent,
};

export type HomeRecentFile = ReturnType<typeof useHomeScreen>['recentFiles'][number];
export type HomeVideoItem = VideoCatalogItemDto;

export default useHomeScreen;
