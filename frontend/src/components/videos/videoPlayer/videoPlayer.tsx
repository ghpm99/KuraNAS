import type { VideoFileDto } from '@/service/videoPlayback';
import { ArrowLeft } from 'lucide-react';
import { useEffect, type ReactNode, type RefObject } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
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

        video.addEventListener('timeupdate', updateTime);
        video.addEventListener('loadedmetadata', updateDuration);
        video.addEventListener('ended', handleEnded);

        return () => {
            video.removeEventListener('timeupdate', updateTime);
            video.removeEventListener('loadedmetadata', updateDuration);
            video.removeEventListener('ended', handleEnded);
        };
    }, [onVideoEnded, setCurrentTime, setDuration, videoRef]);

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

                {children ? <div className={styles.controlsLayer}>{children}</div> : null}
            </div>
        </div>
    );
};

export default VideoPlayer;
