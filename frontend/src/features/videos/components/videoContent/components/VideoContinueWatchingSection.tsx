import { Button, LinearProgress } from '@mui/material';
import { Play } from 'lucide-react';
import { type ReactNode } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import { type VideoContinueItemDto } from '@/service/videoPlayback';
import VideoWatchedToggleButton from './VideoWatchedToggleButton';
import styles from '../videoContent.module.css';

type VideoContinueWatchingSectionProps = {
    items: VideoContinueItemDto[];
    onPlayVideo: (videoId: number, playlistId?: number | null) => void;
    action?: ReactNode;
    onSetWatched?: (videoId: number, watched: boolean) => void;
};

const apiBase = `${getApiV1BaseUrl()}/files`;

const getContinueProgressPercent = (item: VideoContinueItemDto) => {
    if (!(item.duration_seconds > 0)) {
        return 0;
    }
    return Math.max(0, Math.min(100, (item.position_seconds / item.duration_seconds) * 100));
};

export default function VideoContinueWatchingSection({
    items,
    onPlayVideo,
    action,
    onSetWatched,
}: VideoContinueWatchingSectionProps) {
    const { t } = useI18n();

    return (
        <section className={styles.sectionBlock}>
            <div className={styles.sectionHeaderRow}>
                <div className={styles.sectionHeader}>
                    <h2>{t('VIDEO_SECTION_CONTINUE')}</h2>
                    <p>{t('VIDEO_SECTION_CONTINUE_DESCRIPTION')}</p>
                </div>
                {action}
            </div>
            {items.length === 0 ? (
                <div className={styles.sectionEmpty}>{t('VIDEO_NO_CONTINUE_WATCHING')}</div>
            ) : (
                <div className={styles.catalogRail}>
                    {items.map((item) => (
                        <article key={item.video.id} className={styles.catalogCard}>
                            <div className={styles.catalogThumb}>
                                <img
                                    loading="lazy"
                                    src={`${apiBase}/video-thumbnail/${item.video.id}?width=480&height=270`}
                                    alt={item.video.name}
                                />
                            </div>
                            <div className={styles.catalogBody}>
                                <div className={styles.catalogMeta}>
                                    <span className={styles.statusBadge}>
                                        {t('VIDEO_CONTINUE_BADGE_RESUME')}
                                    </span>
                                    <span className={styles.catalogPath}>
                                        {item.video.parent_path}
                                    </span>
                                </div>
                                <h3 className={styles.catalogTitle}>{item.video.name}</h3>
                                <LinearProgress
                                    variant="determinate"
                                    value={getContinueProgressPercent(item)}
                                    aria-label={t('VIDEO_CONTINUE_PROGRESS_LABEL', {
                                        name: item.video.name,
                                    })}
                                />
                                <div className={styles.catalogFooter}>
                                    <p className={styles.catalogFormat}>
                                        {(item.video.format ?? '').toUpperCase()}
                                    </p>
                                    <Button
                                        variant="contained"
                                        size="small"
                                        startIcon={<Play size={14} />}
                                        onClick={() => onPlayVideo(item.video.id, null)}
                                    >
                                        {t('VIDEO_PLAY')}
                                    </Button>
                                    {onSetWatched && (
                                        <VideoWatchedToggleButton
                                            isWatched={false}
                                            onToggle={() => onSetWatched(item.video.id, true)}
                                        />
                                    )}
                                </div>
                            </div>
                        </article>
                    ))}
                </div>
            )}
        </section>
    );
}
