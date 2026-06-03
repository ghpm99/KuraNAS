import { appRoutes } from '@/app/routes';
import useVideoPlayer from '@/features/videos/hooks/useVideoPlayer/useVideoPlayer';
import useI18n from '@/components/i18n/provider/i18nContext';
import {
    getVideoDetailRoute,
    getVideoDetailSlugFromPath,
    getVideoSectionForPlaylist,
    getVideoSectionFromPath,
    getVideoSectionMeta,
} from '@/components/videos/navigation';
import {
    buildVideoPlaylistDetail,
    type VideoDetailItem,
} from '@/components/videos/videoContent/useVideoPlaylistDetail';
import { useSettings } from '@/components/providers/settingsProvider/settingsContext';
import { useCallback, useEffect, useMemo, useRef } from 'react';
import { useLocation, useNavigate, useParams, useSearchParams } from 'react-router-dom';

type VideoPlayerLocationState = {
    from?: string;
    playlistId?: number | null;
};

const slugify = (value: string) =>
    value
        .normalize('NFD')
        .replace(/[\u0300-\u036f]/g, '')
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/^-+|-+$/g, '');

const formatPlaybackTime = (time: number) => {
    if (!Number.isFinite(time) || time <= 0) {
        return '0:00';
    }

    const hours = Math.floor(time / 3600);
    const minutes = Math.floor((time % 3600) / 60);
    const seconds = Math.floor(time % 60);

    if (hours > 0) {
        return `${hours}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
    }

    return `${minutes}:${String(seconds).padStart(2, '0')}`;
};

const getPathname = (value?: string) => {
    if (!value) {
        return '';
    }

    return value.split('?')[0] ?? value;
};

export default function useVideoPlayerScreen() {
    const { t } = useI18n();
    const { settings } = useSettings();
    const { id } = useParams();
    const navigate = useNavigate();
    const location = useLocation();
    const [searchParams] = useSearchParams();
    const locationState = (location.state as VideoPlayerLocationState | null) ?? null;
    const videoId = typeof id === 'string' ? id : '';

    const playlistIdFromUrl = searchParams.get('playlist');
    const resolvedPlaylistId = playlistIdFromUrl
        ? Number(playlistIdFromUrl)
        : (locationState?.playlistId ?? null);

    const {
        videoRef,
        playVideo,
        seekTo,
        setVolume,
        setPlaybackRate,
        toggleFullscreen,
        togglePlayPause,
        nextVideo,
        previousVideo,
        status,
        currentTime,
        duration,
        volume,
        playbackRate,
        isFullscreen,
        setCurrentTime,
        setDuration,
        currentVideo,
        playlist,
        playbackState,
        onVideoEnded,
    } = useVideoPlayer({
        videoId,
        playlistId: resolvedPlaylistId,
        persistProgress: settings.players.remember_video_progress,
    });
    const syncedRouteVideoIdRef = useRef<string | null>(null);

    const fallbackContextRoute = useMemo(() => {
        if (!playlist) {
            return appRoutes.videos;
        }

        return getVideoDetailRoute(getVideoSectionForPlaylist(playlist), slugify(playlist.name));
    }, [playlist]);

    const fromParam = searchParams.get('from');
    const originRoute = fromParam ?? locationState?.from ?? fallbackContextRoute;
    const originPathname = getPathname(originRoute);

    const originLabels = useMemo(() => {
        if (originPathname === appRoutes.home) {
            return {
                badge: t('HOME'),
                context: t('HOME'),
            };
        }

        if (originPathname === appRoutes.files) {
            return {
                badge: t('FILES'),
                context: t('FILES'),
            };
        }

        if (
            originPathname === appRoutes.favorites ||
            originPathname === appRoutes.legacyFavorites
        ) {
            return {
                badge: t('FAVORITES_PAGE_TITLE'),
                context: t('FAVORITES_PAGE_TITLE'),
            };
        }

        if (originPathname.startsWith(appRoutes.videos)) {
            const section = getVideoSectionFromPath(originPathname);
            const sectionLabel = t(getVideoSectionMeta(section).labelKey);
            const isDetailRoute = Boolean(getVideoDetailSlugFromPath(originPathname));

            return {
                badge: isDetailRoute ? sectionLabel : t('NAV_VIDEOS'),
                context: isDetailRoute && playlist?.name ? playlist.name : sectionLabel,
            };
        }

        return {
            badge: t('NAV_VIDEOS'),
            context: playlist?.name ?? t('NAV_VIDEOS'),
        };
    }, [originPathname, playlist?.name, t]);

    const playlistDetail = useMemo(
        () => (playlist ? buildVideoPlaylistDetail(playlist) : null),
        [playlist]
    );
    const orderedItems = useMemo(() => playlistDetail?.orderedItems ?? [], [playlistDetail]);
    const currentVideoId = currentVideo?.id ?? null;
    const currentIndex = orderedItems.findIndex((item) => item.video.id === currentVideoId);
    const hasPreviousVideo = currentIndex > 0;
    const hasNextVideo = currentIndex >= 0 && currentIndex < orderedItems.length - 1;
    const nextItem = hasNextVideo ? (orderedItems[currentIndex + 1] ?? null) : null;
    const nextItemId = nextItem?.video.id ?? null;

    const relatedItems = useMemo(() => {
        if (!orderedItems.length || !currentVideoId) {
            return [] as VideoDetailItem[];
        }

        const itemsAfterCurrent =
            currentIndex >= 0 ? orderedItems.slice(currentIndex + 1) : orderedItems;
        const itemsBeforeCurrent = currentIndex > 0 ? orderedItems.slice(0, currentIndex) : [];

        return [...itemsAfterCurrent, ...itemsBeforeCurrent]
            .filter((item) => item.video.id !== currentVideoId && item.video.id !== nextItemId)
            .slice(0, 4);
    }, [currentIndex, currentVideoId, nextItemId, orderedItems]);

    const positionLabel =
        currentIndex >= 0 && orderedItems.length > 1
            ? t('VIDEO_PLAYER_POSITION', {
                  current: String(currentIndex + 1),
                  total: String(orderedItems.length),
              })
            : '';

    const resumeLabel =
        playbackState && playbackState.current_time > 0 && !playbackState.completed
            ? t('VIDEO_PLAYER_RESUME_POSITION', {
                  time: formatPlaybackTime(playbackState.current_time),
              })
            : '';

    const metadataLine = [
        playlist?.name && playlist.name !== currentVideo?.name ? playlist.name : '',
        positionLabel,
        resumeLabel,
    ]
        .filter(Boolean)
        .join(' • ');

    const relatedTitle =
        playlistDetail?.hasEpisodeData ||
        playlist?.classification === 'series' ||
        playlist?.classification === 'anime'
            ? t('VIDEO_PLAYER_NEXT_EPISODES')
            : t('VIDEO_PLAYER_RELATED_VIDEOS');

    const resolvedPlaylistIdForState =
        playbackState?.playlist_id ?? playlist?.id ?? resolvedPlaylistId ?? null;

    const buildVideoPlayerUrl = useCallback(
        (targetVideoId: number | string) => {
            const params = new URLSearchParams();
            if (resolvedPlaylistIdForState) {
                params.set('playlist', String(resolvedPlaylistIdForState));
            }
            if (originRoute) {
                params.set('from', originRoute);
            }
            const qs = params.toString();
            return `${appRoutes.videoPlayerBase}/${targetVideoId}${qs ? `?${qs}` : ''}`;
        },
        [originRoute, resolvedPlaylistIdForState]
    );

    const persistedState = useMemo(
        () => ({
            ...(locationState ?? {}),
            from: originRoute,
            playlistId: resolvedPlaylistIdForState,
        }),
        [locationState, originRoute, resolvedPlaylistIdForState]
    );

    const openVideo = useCallback(
        (targetVideoId: number) => {
            navigate(buildVideoPlayerUrl(targetVideoId), {
                state: persistedState,
            });
        },
        [buildVideoPlayerUrl, navigate, persistedState]
    );

    const handleBack = useCallback(() => {
        navigate(originRoute);
    }, [navigate, originRoute]);

    const handlePlaybackEnded = useCallback(async () => {
        await onVideoEnded();
        if (settings.players.autoplay_next_video && hasNextVideo) {
            await nextVideo();
        }
    }, [hasNextVideo, nextVideo, onVideoEnded, settings.players.autoplay_next_video]);

    useEffect(() => {
        if (!videoId) {
            return;
        }

        if (syncedRouteVideoIdRef.current === videoId) {
            syncedRouteVideoIdRef.current = null;
            return;
        }

        playVideo();
    }, [playVideo, videoId]);

    useEffect(() => {
        if (!currentVideo?.id || String(currentVideo.id) === videoId) {
            return;
        }

        syncedRouteVideoIdRef.current = String(currentVideo.id);
        navigate(buildVideoPlayerUrl(currentVideo.id), {
            replace: true,
            state: persistedState,
        });
    }, [buildVideoPlayerUrl, currentVideo?.id, navigate, persistedState, videoId]);

    return {
        videoId,
        isInvalidVideoId: !videoId,
        handleBack,
        handlePlaybackEnded,
        openVideo,
        currentVideo,
        contextTitle: playlist?.name ?? currentVideo?.name ?? t('NAV_VIDEOS'),
        originBadgeLabel: originLabels.badge,
        contextDescription: t('VIDEO_PLAYER_FROM_CONTEXT', {
            context: originLabels.context,
        }),
        metadataLine,
        nextItem,
        relatedItems,
        relatedTitle,
        hasNextVideo,
        hasPreviousVideo,
        videoRef,
        seekTo,
        setVolume,
        setPlaybackRate,
        toggleFullscreen,
        togglePlayPause,
        nextVideo,
        previousVideo,
        status,
        currentTime,
        duration,
        volume,
        playbackRate,
        isFullscreen,
        setCurrentTime,
        setDuration,
    };
}
