import { ArrowLeft, Play } from 'lucide-react';
import { useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import { type VideoPlaylistDto } from '@/service/videoPlayback';
import { useVideoPlaylistDetail } from '../useVideoPlaylistDetail';
import VideoDetailListItem from './VideoDetailListItem';
import styles from '../videoContent.module.css';

type VideoSeriesDetailViewProps = {
    playlist: VideoPlaylistDto;
    onBack: () => void;
    onOpenVideo: (videoId: number) => void;
};

const apiBase = `${getApiV1BaseUrl()}/files`;

export default function VideoSeriesDetailView({
    playlist,
    onBack,
    onOpenVideo,
}: VideoSeriesDetailViewProps) {
    const { t } = useI18n();
    const { groupedSeasons, completedCount, orderedItems, resumeItem } =
        useVideoPlaylistDetail(playlist);
    const [selectedSeasonKey, setSelectedSeasonKey] = useState('');
    const resolvedSeasonKey =
        groupedSeasons.find((group) => group.key === selectedSeasonKey)?.key ??
        groupedSeasons[0]?.key ??
        '';

    const visibleSeasons =
        groupedSeasons.length > 0
            ? groupedSeasons.filter((group) => group.key === resolvedSeasonKey)
            : [
                  {
                      key: 'season-1',
                      label: '1',
                      items: orderedItems,
                  },
              ];
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
                    <p className={styles.heroEyebrow}>{t('VIDEO_DETAIL_SERIES_EYEBROW')}</p>
                    <h1 className={styles.heroTitle}>{playlist.name}</h1>
                    <p className={styles.heroMeta}>
                        {t('VIDEO_DETAIL_SERIES_PROGRESS', {
                            completed: String(completedCount),
                            count: String(orderedItems.length),
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

            {groupedSeasons.length > 1 && (
                <section className={styles.sectionBlock}>
                    <div className={styles.seasonSelector}>
                        {groupedSeasons.map((group) => (
                            <button
                                key={group.key}
                                type="button"
                                className={
                                    group.key === resolvedSeasonKey
                                        ? styles.seasonChipActive
                                        : styles.seasonChip
                                }
                                onClick={() => setSelectedSeasonKey(group.key)}
                            >
                                {t('VIDEO_DETAIL_SEASON_LABEL', { season: group.label })}
                            </button>
                        ))}
                    </div>
                </section>
            )}

            {visibleSeasons.map((group) => (
                <section key={group.key} className={styles.sectionBlock}>
                    <div className={styles.sectionHeader}>
                        <h2>{t('VIDEO_DETAIL_SEASON_LABEL', { season: group.label })}</h2>
                        <p>
                            {t('VIDEO_DETAIL_SEASON_DESCRIPTION', {
                                count: String(group.items.length),
                            })}
                        </p>
                    </div>
                    <div className={styles.detailListStack}>
                        {group.items.map((item) => (
                            <VideoDetailListItem
                                key={item.id}
                                item={item}
                                onOpenVideo={onOpenVideo}
                            />
                        ))}
                    </div>
                </section>
            ))}
        </div>
    );
}
