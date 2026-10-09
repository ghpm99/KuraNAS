import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import type { IImageData } from '@/types/image';
import type { IMusicData } from '@/types/music';
import {
    fetchAnalyticsHealth,
    fetchAnalyticsRecentFiles,
    fetchAnalyticsStorage,
} from '@/service/analytics';
import { getTrackDurationSeconds } from '@/utils/music';
import { getStarredFiles } from '@/service/files';
import { getImageFiles } from '@/service/image';
import { getPlayerQueue, getPlayerState } from '@/service/playerState';
import { queueEntryToTrack } from '@/features/music/components/musicQueueTracks';
import {
    getVideoHomeCatalog,
    getVideoPlaybackState,
    type VideoCatalogItemDto,
    type VideoFileDto,
} from '@/service/videoPlayback';
import { useQuery } from '@tanstack/react-query';
import { useMemo } from 'react';
import { analyticsStaleTimeMs } from '@/components/providers/queryFreshness';

const homeAnalyticsPeriod = '30d' as const;
const videoHomeLimit = 12;
const homeFavoritesLimit = 6;
const homeImagesLimit = 6;
const homeRecentFilesLimit = 6;

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
    const { queue, currentTrack, currentTime, duration, isPlaying } = useGlobalMusic();

    const analyticsQuery = useQuery({
        queryKey: ['home', 'analytics', homeAnalyticsPeriod],
        queryFn: async () => {
            const [storageStats, health, recentFiles] = await Promise.all([
                fetchAnalyticsStorage(homeAnalyticsPeriod),
                fetchAnalyticsHealth(),
                fetchAnalyticsRecentFiles(homeRecentFilesLimit),
            ]);
            return {
                storage: storageStats.storage,
                counts: storageStats.counts,
                health,
                recent_files: recentFiles,
            };
        },
        staleTime: analyticsStaleTimeMs,
        retry: false,
    });
    const favoritesQuery = useQuery({
        queryKey: ['home', 'favorites'],
        queryFn: () =>
            getStarredFiles({
                page: 1,
                pageSize: homeFavoritesLimit,
            }),
    });
    const imagesQuery = useQuery({
        queryKey: ['home', 'images'],
        queryFn: () => getImageFiles(1, homeImagesLimit, 'date'),
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

    const playerQueueQuery = useQuery({
        queryKey: ['home', 'music-player-queue'],
        queryFn: () => getPlayerQueue(),
        retry: false,
    });

    const recentFiles = analyticsQuery.data?.recent_files?.slice(0, homeRecentFilesLimit) ?? [];
    const favoriteItems = favoritesQuery.data?.items?.slice(0, homeFavoritesLimit) ?? [];
    const recentImages = imagesQuery.data?.items?.slice(0, homeImagesLimit) ?? [];

    const videoContinueItems = useMemo(() => {
        const continueSection = videoCatalogQuery.data?.sections.find(
            (section) => section.key === 'continue'
        );
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
                session.playback_state.duration
            ),
            playlistId: session.playback_state.playlist_id,
        };
    }, [videoPlaybackQuery.data]);

    const fallbackMusicTrack = useMemo(() => {
        const savedEntry = playerQueueQuery.data?.items?.[playerQueueQuery.data.current_index];
        return savedEntry ? queueEntryToTrack(savedEntry) : null;
    }, [playerQueueQuery.data]);

    const activeTrackDurationSeconds = getTrackDurationSeconds(
        (currentTrack ?? fallbackMusicTrack)?.metadata
    );

    const musicResume = useMemo<HomeMusicResume | null>(() => {
        const activeTrack = currentTrack ?? fallbackMusicTrack;
        if (!activeTrack) {
            return null;
        }

        const progressSeconds = currentTrack
            ? currentTime
            : (playerStateQuery.data?.current_position ?? 0);
        const durationSeconds = currentTrack
            ? Math.max(duration, activeTrackDurationSeconds)
            : activeTrackDurationSeconds;
        const queueCount = queue.length || playerQueueQuery.data?.items?.length || 0;

        return {
            track: activeTrack,
            progressSeconds,
            durationSeconds,
            progressPercent: getProgressPercent(progressSeconds, durationSeconds),
            queueCount,
            isPlaying: currentTrack ? isPlaying : false,
        };
    }, [
        activeTrackDurationSeconds,
        currentTime,
        currentTrack,
        duration,
        fallbackMusicTrack,
        isPlaying,
        playerQueueQuery.data?.items?.length,
        playerStateQuery.data?.current_position,
        queue.length,
    ]);

    return {
        recentFiles,
        favoriteItems,
        recentImages,
        videoContinueItems,
        videoResume,
        musicResume,
        analytics: analyticsQuery.data
            ? {
                  storage: analyticsQuery.data.storage,
                  counts: analyticsQuery.data.counts,
                  health: analyticsQuery.data.health,
              }
            : null,
        isAnalyticsLoading: analyticsQuery.isLoading,
        isFavoritesLoading: favoritesQuery.isLoading,
        isImagesLoading: imagesQuery.isLoading,
        isVideoLoading: videoCatalogQuery.isLoading || videoPlaybackQuery.isLoading,
        isMusicLoading: playerStateQuery.isLoading || playerQueueQuery.isLoading,
    };
};

export const homeScreenUtils = {
    getProgressPercent,
};

export type HomeRecentFile = ReturnType<typeof useHomeScreen>['recentFiles'][number];
export type HomeFavoriteFile = FileData;
export type HomeRecentImage = IImageData;
export type HomeVideoItem = VideoCatalogItemDto;

export default useHomeScreen;
