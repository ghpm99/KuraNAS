import { Check, Play } from 'lucide-react';
import { Typography } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import type { VideoPlaylistItemDto } from '@/service/videoPlayback';
import { parseEpisode } from '../useVideoPlaylistDetail';
import VideoStarButton from '../../VideoStarButton';
import VideoCorrectionMenu from './VideoCorrectionMenu';
import VideoWatchedToggleButton from './VideoWatchedToggleButton';
import { formatVideoDuration } from './formatVideoDuration';
import styles from '../videoContent.module.css';

type VideoEpisodeRowProps = {
    episode: VideoPlaylistItemDto;
    onOpenVideo: (videoId: number) => void;
    onSetWatched?: (videoId: number, watched: boolean) => void;
};

const thumbnailBase = `${getApiV1BaseUrl()}/files`;

export default function VideoEpisodeRow({
    episode,
    onOpenVideo,
    onSetWatched,
}: VideoEpisodeRowProps) {
    const { t } = useI18n();
    const isWatched = episode.status === 'completed';
    const duration = formatVideoDuration(episode.video.duration_seconds);
    const hasProgress = episode.progress_pct > 0 && episode.status !== 'not_started';

    return (
        <div className={styles.watchedRow}>
            <button
                type="button"
                className={styles.detailListItem}
                onClick={() => onOpenVideo(episode.video.id)}
            >
                <div className={styles.detailListThumb}>
                    <img
                        loading="lazy"
                        src={`${thumbnailBase}/video-thumbnail/${episode.video.id}?width=320&height=180`}
                        alt={episode.video.name}
                    />
                </div>
                <div className={styles.detailListBody}>
                    <div className={styles.detailListHeader}>
                        {episode.episode != null && (
                            <span className={styles.sequenceTag}>
                                {t('VIDEO_SERIES_EPISODE_NUMBER', {
                                    episode: String(episode.episode),
                                })}
                            </span>
                        )}
                        {isWatched && (
                            <span className={styles.statusBadge}>
                                <Check size={12} aria-hidden="true" /> {t('VIDEO_STATUS_COMPLETED')}
                            </span>
                        )}
                    </div>
                    <Typography className={styles.detailListTitle}>
                        {parseEpisode(episode.video.name).displayTitle || episode.video.name}
                    </Typography>
                    {duration && (
                        <Typography className={styles.detailListMeta}>{duration}</Typography>
                    )}
                    {hasProgress && (
                        <div
                            className={styles.progressTrack}
                            role="progressbar"
                            aria-valuenow={Math.round(episode.progress_pct)}
                            aria-valuemin={0}
                            aria-valuemax={100}
                        >
                            <div
                                className={styles.progressFill}
                                style={{ width: `${episode.progress_pct}%` }}
                            />
                        </div>
                    )}
                </div>
                <div className={styles.detailPlay}>
                    <Play size={16} />
                </div>
            </button>
            <VideoStarButton
                videoId={episode.video.id}
                videoName={episode.video.name}
                isStarred={episode.video.starred}
            />
            <VideoCorrectionMenu
                fileId={episode.video.id}
                fileName={episode.video.name}
                canCorrectEpisode
                season={episode.season}
                episode={episode.episode}
            />
            {onSetWatched && (
                <VideoWatchedToggleButton
                    isWatched={isWatched}
                    onToggle={() => onSetWatched(episode.video.id, !isWatched)}
                />
            )}
        </div>
    );
}
