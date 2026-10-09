import type { VideoFileDto } from '@/service/videoPlayback';
import { getFileDownloadUrl } from '@/service/files';
import { ArrowLeft } from 'lucide-react';
import { useEffect, type ReactNode, type RefObject } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import {
    getPlaybackErrorKindFromMediaErrorCode,
    PLAYBACK_ERROR_MESSAGE_KEYS,
    type PlaybackErrorKind,
} from './playbackError';
import styles from './VideoPlayer.module.css';

interface VideoPlayerProps {
    currentVideo: VideoFileDto | null;
    videoRef: RefObject<HTMLVideoElement | null>;
    setCurrentTime: (time: number) => void;
    setDuration: (duration: number) => void;
    onBack: () => void;
    onVideoEnded: () => void | Promise<void>;
    originBadgeLabel: string;
    contextDescription: string;
    metadataLine: string;
    playbackError?: PlaybackErrorKind | null;
    onPlaybackError?: (errorKind: PlaybackErrorKind) => void;
    onRetryPlayback?: () => void;
    children?: ReactNode;
}

const VideoPlayer = ({
    currentVideo,
    videoRef,
    setCurrentTime,
    setDuration,
    onBack,
    onVideoEnded,
    originBadgeLabel,
    contextDescription,
    metadataLine,
    playbackError = null,
    onPlaybackError,
    onRetryPlayback,
    children,
}: VideoPlayerProps) => {
    const { t } = useI18n();

    useEffect(() => {
        const video = videoRef.current;
        if (!video) {
            return;
        }

        const updateTime = () => setCurrentTime(video.currentTime);
        const updateDuration = () => setDuration(video.duration);
        const handleEnded = () => {
            void onVideoEnded();
        };

        const handleError = () => {
            onPlaybackError?.(getPlaybackErrorKindFromMediaErrorCode(video.error?.code));
        };

        video.addEventListener('timeupdate', updateTime);
        video.addEventListener('loadedmetadata', updateDuration);
        video.addEventListener('ended', handleEnded);
        video.addEventListener('error', handleError);

        return () => {
            video.removeEventListener('timeupdate', updateTime);
            video.removeEventListener('loadedmetadata', updateDuration);
            video.removeEventListener('ended', handleEnded);
            video.removeEventListener('error', handleError);
        };
    }, [onPlaybackError, onVideoEnded, setCurrentTime, setDuration, videoRef]);

    return (
        <div className={styles.player}>
            <div className={styles.container}>
                <video ref={videoRef} className={styles.video} preload="metadata" playsInline />
                <div className={styles.overlay}>
                    <div className={styles.header}>
                        <button type="button" className={styles.backButton} onClick={onBack}>
                            <ArrowLeft size={16} />
                            <span>{t('VIDEO_BACK')}</span>
                        </button>
                        <span className={styles.contextBadge}>{originBadgeLabel}</span>
                    </div>

                    <div className={styles.info}>
                        <p className={styles.contextDescription}>{contextDescription}</p>
                        <h1 className={styles.title}>
                            {currentVideo?.name ?? t('VIDEO_NO_VIDEO_PLAYING')}
                        </h1>
                        {metadataLine ? <p className={styles.metadata}>{metadataLine}</p> : null}
                    </div>
                </div>

                {playbackError ? (
                    <div className={styles.errorOverlay} role="alert">
                        <p className={styles.errorMessage}>
                            {t(PLAYBACK_ERROR_MESSAGE_KEYS[playbackError])}
                        </p>
                        {currentVideo ? (
                            <p className={styles.errorFileName}>{currentVideo.name}</p>
                        ) : null}
                        <div className={styles.errorActions}>
                            {currentVideo ? (
                                <a
                                    className={styles.errorAction}
                                    href={getFileDownloadUrl(currentVideo.id)}
                                    download
                                >
                                    {t('VIDEO_ERROR_DOWNLOAD_ORIGINAL')}
                                </a>
                            ) : null}
                            <button
                                type="button"
                                className={styles.errorAction}
                                onClick={onRetryPlayback}
                            >
                                {t('VIDEO_ERROR_RETRY')}
                            </button>
                        </div>
                    </div>
                ) : null}

                {children ? <div className={styles.controlsLayer}>{children}</div> : null}
            </div>
        </div>
    );
};

export default VideoPlayer;
