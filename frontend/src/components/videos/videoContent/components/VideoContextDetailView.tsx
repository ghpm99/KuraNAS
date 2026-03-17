import { ArrowLeft, Play } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import { type VideoPlaylistDto } from '@/service/videoPlayback';
import { useVideoPlaylistDetail } from '../useVideoPlaylistDetail';
import VideoDetailListItem from './VideoDetailListItem';
import styles from '../videoContent.module.css';

type VideoContextDetailViewProps = {
    playlist: VideoPlaylistDto;
    onBack: () => void;
    onOpenVideo: (videoId: number) => void;
};

const apiBase = `${getApiV1BaseUrl()}/files`;

export default function VideoContextDetailView({
    playlist,
    onBack,
    onOpenVideo,
}: VideoContextDetailViewProps) {
    const { t } = useI18n();
    const { orderedItems, completedCount, resumeItem } = useVideoPlaylistDetail(playlist);
    const coverId = resumeItem?.video.id ?? playlist.cover_video_id ?? orderedItems[0]?.video.id;

    return (
        <div className={styles.page}>
            <section className={styles.hero}>
                {coverId && (
                    <img
                        className={styles.heroImage}
                        src={`${apiBase}/video-thumbnail/${coverId}?width=1280&height=720`}
                        alt={playlist.name}
                    />
                )}
                <div className={styles.heroShade} />
                <div className={styles.heroContent}>
                    <button type="button" className={styles.backBtn} onClick={onBack}>
                        <ArrowLeft size={16} />
                        <span>{t('VIDEO_BACK_TO_VIDEOS')}</span>
                    </button>
                    <p className={styles.heroEyebrow}>
                        {(playlist.classification || 'personal').toUpperCase()}
                    </p>
                    <h1 className={styles.heroTitle}>{playlist.name}</h1>
                    <p className={styles.heroMeta}>
                        {t('VIDEO_DETAIL_COLLECTION_META', {
                            count: String(playlist.item_count),
                        })}
                    </p>
                    <div className={styles.detailHeroActions}>
                        <button
                            type="button"
                            className={styles.actionBtnPrimary}
                            onClick={() => resumeItem && onOpenVideo(resumeItem.video.id)}
                            disabled={!resumeItem}
                        >
                            <Play size={16} />
                            <span>
                                {resumeItem?.status === 'in_progress'
                                    ? t('VIDEO_DETAIL_RESUME_ACTION')
                                    : t('VIDEO_PLAY')}
                            </span>
                        </button>
                    </div>
                </div>
            </section>

            <section className={styles.detailStatsGrid}>
                <div className={styles.detailStatCard}>
                    <span className={styles.detailStatLabel}>
                        {t('VIDEO_DETAIL_COLLECTION_TOTAL')}
                    </span>
                    <strong>{orderedItems.length}</strong>
                </div>
                <div className={styles.detailStatCard}>
                    <span className={styles.detailStatLabel}>
                        {t('VIDEO_DETAIL_COLLECTION_COMPLETED')}
                    </span>
                    <strong>{completedCount}</strong>
                </div>
                <div className={styles.detailStatCard}>
                    <span className={styles.detailStatLabel}>
                        {t('VIDEO_DETAIL_COLLECTION_PENDING')}
                    </span>
                    <strong>{Math.max(orderedItems.length - completedCount, 0)}</strong>
                </div>
            </section>

            <section className={styles.sectionBlock}>
                <div className={styles.sectionHeader}>
                    <h2>{t('VIDEO_DETAIL_COLLECTION_ITEMS')}</h2>
                    <p>{t('VIDEO_DETAIL_COLLECTION_ITEMS_DESCRIPTION')}</p>
                </div>
                <div className={styles.detailListStack}>
                    {orderedItems.map((item) => (
                        <VideoDetailListItem key={item.id} item={item} onOpenVideo={onOpenVideo} />
                    ))}
                </div>
            </section>
        </div>
    );
}
