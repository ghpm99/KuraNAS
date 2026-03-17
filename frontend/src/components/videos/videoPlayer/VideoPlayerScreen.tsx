import useI18n from '@/components/i18n/provider/i18nContext';
import VideoControls from '@/components/videos/videoControls/videoControls';
import VideoPlayer from '@/components/videos/videoPlayer/videoPlayer';
import styles from './VideoPlayerScreen.module.css';
import useVideoPlayerScreen from './useVideoPlayerScreen';

const statusKeyMap = {
    not_started: 'VIDEO_STATUS_NOT_STARTED',
    in_progress: 'VIDEO_STATUS_IN_PROGRESS',
    completed: 'VIDEO_STATUS_COMPLETED',
} as const;

export default function VideoPlayerScreen() {
    const { t } = useI18n();
    const {
        isInvalidVideoId,
        handleBack,
        handlePlaybackEnded,
        openVideo,
        currentVideo,
        contextTitle,
        originBadgeLabel,
        contextDescription,
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
    } = useVideoPlayerScreen();

    if (isInvalidVideoId) {
        return <div>{t('VIDEO_INVALID_ID')}</div>;
    }

    return (
        <div className={styles.page}>
            <section className={styles.playerColumn}>
                <VideoPlayer
                    currentVideo={currentVideo}
                    videoRef={videoRef}
                    setCurrentTime={setCurrentTime}
                    setDuration={setDuration}
                    onBack={handleBack}
                    onVideoEnded={handlePlaybackEnded}
                    originBadgeLabel={originBadgeLabel}
                    contextDescription={contextDescription}
                    metadataLine={metadataLine}
                >
                    <VideoControls
                        currentTime={currentTime}
                        duration={duration}
                        isFullscreen={isFullscreen}
                        isPlaying={status === 'playing'}
                        volume={volume}
                        playbackRate={playbackRate}
                        seekTo={seekTo}
                        setVolume={setVolume}
                        setPlaybackRate={setPlaybackRate}
                        toggleFullscreen={toggleFullscreen}
                        togglePlayPause={togglePlayPause}
                        nextVideo={nextVideo}
                        previousVideo={previousVideo}
                        canGoNext={hasNextVideo}
                        canGoPrevious={hasPreviousVideo}
                    />
                </VideoPlayer>
            </section>

            <aside className={styles.contextPanel}>
                <header className={styles.contextHeader}>
                    <p className={styles.contextEyebrow}>{originBadgeLabel}</p>
                    <h2 className={styles.contextTitle}>{contextTitle}</h2>
                    <p className={styles.contextDescription}>{contextDescription}</p>
                </header>

                {nextItem ? (
                    <button
                        type="button"
                        className={styles.upNextCard}
                        onClick={() => openVideo(nextItem.video.id)}
                        aria-label={t('VIDEO_PLAYER_OPEN_VIDEO', {
                            name: nextItem.displayTitle || nextItem.video.name,
                        })}
                    >
                        <span className={styles.cardEyebrow}>{t('VIDEO_PLAYER_UP_NEXT')}</span>
                        <strong className={styles.cardTitle}>
                            {nextItem.displayTitle || nextItem.video.name}
                        </strong>
                        <div className={styles.cardMeta}>
                            {nextItem.sequenceLabel ? <span>{nextItem.sequenceLabel}</span> : null}
                            <span>{t(statusKeyMap[nextItem.status])}</span>
                        </div>
                        {nextItem.progress_pct > 0 ? (
                            <div className={styles.progressTrack} aria-hidden="true">
                                <div
                                    className={styles.progressFill}
                                    style={{ width: `${nextItem.progress_pct}%` }}
                                />
                            </div>
                        ) : null}
                    </button>
                ) : null}

                {relatedItems.length > 0 ? (
                    <section className={styles.relatedSection}>
                        <div className={styles.relatedHeader}>
                            <h3>{relatedTitle}</h3>
                        </div>

                        <div className={styles.relatedList}>
                            {relatedItems.map((item) => (
                                <button
                                    key={item.id}
                                    type="button"
                                    className={styles.relatedItem}
                                    onClick={() => openVideo(item.video.id)}
                                    aria-label={t('VIDEO_PLAYER_OPEN_VIDEO', {
                                        name: item.displayTitle || item.video.name,
                                    })}
                                >
                                    <div className={styles.relatedBody}>
                                        <div className={styles.relatedHeaderRow}>
                                            {item.sequenceLabel ? (
                                                <span className={styles.sequenceTag}>
                                                    {item.sequenceLabel}
                                                </span>
                                            ) : null}
                                            <span className={styles.statusBadge}>
                                                {t(statusKeyMap[item.status])}
                                            </span>
                                        </div>
                                        <strong className={styles.relatedTitle}>
                                            {item.displayTitle || item.video.name}
                                        </strong>
                                        <span className={styles.relatedMeta}>
                                            {item.video.name}
                                        </span>
                                        {item.progress_pct > 0 ? (
                                            <div
                                                className={styles.progressTrack}
                                                aria-hidden="true"
                                            >
                                                <div
                                                    className={styles.progressFill}
                                                    style={{ width: `${item.progress_pct}%` }}
                                                />
                                            </div>
                                        ) : null}
                                    </div>
                                </button>
                            ))}
                        </div>
                    </section>
                ) : null}
            </aside>
        </div>
    );
}
