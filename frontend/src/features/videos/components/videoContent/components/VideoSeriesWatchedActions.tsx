import { Eye, EyeOff } from 'lucide-react';
import { useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useVideoSeriesWatchedMutation } from '@/features/videos/providers/videoContentProvider/useVideoSeriesWatchedMutation';
import type { VideoSeriesSeasonDto } from '@/service/videoPlayback';
import VideoConfirmDialog from './VideoConfirmDialog';
import styles from '../videoContent.module.css';

type VideoSeriesWatchedActionsProps = {
    playlistId: number;
    playlistName: string;
    seasons: VideoSeriesSeasonDto[];
    selectedSeason?: number;
};

const isFullyWatched = (seasons: VideoSeriesSeasonDto[]) =>
    seasons.length > 0 &&
    seasons.every(
        (season) => season.episode_count > 0 && season.watched_count >= season.episode_count
    );

export default function VideoSeriesWatchedActions({
    playlistId,
    playlistName,
    seasons,
    selectedSeason,
}: VideoSeriesWatchedActionsProps) {
    const { t } = useI18n();
    const { markWatched, isMarkingWatched } = useVideoSeriesWatchedMutation(playlistId);
    const [isSeriesConfirmOpen, setIsSeriesConfirmOpen] = useState(false);

    if (seasons.length === 0) return null;

    const season = seasons.find((candidate) => candidate.season === selectedSeason);
    const isSeasonWatched = season ? isFullyWatched([season]) : false;
    const isSeriesWatched = isFullyWatched(seasons);
    const seriesConfirmKeyPrefix = isSeriesWatched
        ? 'VIDEO_SERIES_MARK_UNWATCHED_CONFIRM'
        : 'VIDEO_SERIES_MARK_WATCHED_CONFIRM';

    return (
        <>
            {selectedSeason !== undefined && (
                <button
                    type="button"
                    className={styles.actionBtn}
                    disabled={isMarkingWatched}
                    onClick={() =>
                        markWatched({ watched: !isSeasonWatched, season: selectedSeason })
                    }
                >
                    {isSeasonWatched ? <EyeOff size={14} /> : <Eye size={14} />}
                    {t(
                        isSeasonWatched
                            ? 'VIDEO_SERIES_MARK_SEASON_UNWATCHED'
                            : 'VIDEO_SERIES_MARK_SEASON_WATCHED'
                    )}
                </button>
            )}
            <button
                type="button"
                className={styles.actionBtn}
                disabled={isMarkingWatched}
                onClick={() => setIsSeriesConfirmOpen(true)}
            >
                {isSeriesWatched ? <EyeOff size={14} /> : <Eye size={14} />}
                {t(
                    isSeriesWatched
                        ? 'VIDEO_SERIES_MARK_SERIES_UNWATCHED'
                        : 'VIDEO_SERIES_MARK_SERIES_WATCHED'
                )}
            </button>
            <VideoConfirmDialog
                isOpen={isSeriesConfirmOpen}
                title={t(`${seriesConfirmKeyPrefix}_TITLE`)}
                message={t(`${seriesConfirmKeyPrefix}_MESSAGE`, { name: playlistName })}
                confirmLabel={t(
                    isSeriesWatched
                        ? 'VIDEO_SERIES_MARK_SERIES_UNWATCHED'
                        : 'VIDEO_SERIES_MARK_SERIES_WATCHED'
                )}
                onConfirm={() => markWatched({ watched: !isSeriesWatched })}
                onClose={() => setIsSeriesConfirmOpen(false)}
            />
        </>
    );
}
