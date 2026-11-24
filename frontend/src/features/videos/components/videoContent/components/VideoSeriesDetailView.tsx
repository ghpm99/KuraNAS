import { ArrowLeft } from 'lucide-react';
import { CircularProgress } from '@mui/material';
import ErrorState from '@/components/errorState/errorState';
import useI18n from '@/components/i18n/provider/i18nContext';
import ListFeedback from '@/components/listFeedback/ListFeedback';
import useRetryFailedPage from '@/shared/hooks/useRetryFailedPage';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { getVideoPlaylistPosterUrl, type VideoPlaylistDto } from '@/service/videoPlayback';
import { useVideoSeriesDetail } from '../useVideoSeriesDetail';
import VideoEpisodeRow from './VideoEpisodeRow';
import VideoHidePlaylistButton from './VideoHidePlaylistButton';
import VideoSeriesResumeButton from './VideoSeriesResumeButton';
import styles from '../videoContent.module.css';

type VideoSeriesDetailViewProps = {
    playlist: Pick<VideoPlaylistDto, 'id' | 'name' | 'item_count'>;
    onBack: () => void;
    onOpenVideo: (videoId: number) => void;
    isChangingVisibility?: boolean;
    onHide?: () => void;
    onSetWatched?: (videoId: number, watched: boolean) => void;
};

const heroPosterWidth = 1280;
const heroPosterHeight = 720;

export default function VideoSeriesDetailView({
    playlist,
    onBack,
    onOpenVideo,
    isChangingVisibility = false,
    onHide,
    onSetWatched,
}: VideoSeriesDetailViewProps) {
    const { t } = useI18n();
    const {
        headerQuery,
        seasons,
        resume,
        selectedSeason,
        selectSeason,
        episodesQuery,
        episodes,
        totalEpisodes,
        watchedEpisodes,
        selectedSeasonEpisodeCount,
    } = useVideoSeriesDetail(playlist.id);
    const retryEpisodes = useRetryFailedPage(episodesQuery);

    if (headerQuery.isError) {
        return (
            <div className={styles.page}>
                <button type="button" className={styles.backBtn} onClick={onBack}>
                    <ArrowLeft size={16} />
                    <span>{t('VIDEO_BACK_TO_VIDEOS')}</span>
                </button>
                <ErrorState
                    title={t('VIDEO_SECTION_LOAD_ERROR', { section: playlist.name })}
                    backendMessage={extractBackendErrorMessage(headerQuery.error)}
                    onRetry={() => void headerQuery.refetch()}
                />
            </div>
        );
    }

    const episodeTotal = totalEpisodes || playlist.item_count;
    const progressPercent = episodeTotal > 0 ? (watchedEpisodes / episodeTotal) * 100 : 0;
    const isLoadingEpisodes = headerQuery.isPending || episodesQuery.isPending;

    return (
        <div className={styles.page}>
            <section className={styles.hero}>
                <img
                    className={styles.heroImage}
                    src={getVideoPlaylistPosterUrl(playlist.id, heroPosterWidth, heroPosterHeight)}
                    alt={playlist.name}
                />
                <div className={styles.heroShade} />
                <div className={styles.heroContent}>
                    <button type="button" className={styles.backBtn} onClick={onBack}>
                        <ArrowLeft size={16} />
                        <span>{t('VIDEO_BACK_TO_VIDEOS')}</span>
                    </button>
                    <p className={styles.heroEyebrow}>{t('VIDEO_DETAIL_SERIES_EYEBROW')}</p>
                    <h2 className={styles.heroTitle}>{playlist.name}</h2>
                    <p className={styles.heroMeta}>
                        {t('VIDEO_DETAIL_SERIES_PROGRESS', {
                            completed: String(watchedEpisodes),
                            count: String(episodeTotal),
                        })}
                    </p>
                    <div
                        className={styles.progressTrack}
                        role="progressbar"
                        aria-valuenow={Math.round(progressPercent)}
                        aria-valuemin={0}
                        aria-valuemax={100}
                    >
                        <div
                            className={styles.progressFill}
                            style={{ width: `${progressPercent}%` }}
                        />
                    </div>
                    <div className={styles.detailHeroActions}>
                        <VideoSeriesResumeButton resume={resume} onOpenVideo={onOpenVideo} />
                        <VideoHidePlaylistButton
                            isChanging={isChangingVisibility}
                            onHide={onHide}
                        />
                    </div>
                </div>
            </section>

            {seasons.length > 1 && (
                <section className={styles.sectionBlock}>
                    <div className={styles.seasonSelector}>
                        {seasons.map((season) => (
                            <button
                                key={season.season}
                                type="button"
                                aria-pressed={season.season === selectedSeason}
                                className={
                                    season.season === selectedSeason
                                        ? styles.seasonChipActive
                                        : styles.seasonChip
                                }
                                onClick={() => selectSeason(season.season)}
                            >
                                {t('VIDEO_DETAIL_SEASON_LABEL', { season: String(season.season) })}
                            </button>
                        ))}
                    </div>
                </section>
            )}

            <section className={styles.sectionBlock}>
                {selectedSeason !== undefined && (
                    <div className={styles.sectionHeader}>
                        <h2>
                            {t('VIDEO_DETAIL_SEASON_LABEL', { season: String(selectedSeason) })}
                        </h2>
                        <p>
                            {t('VIDEO_DETAIL_SEASON_DESCRIPTION', {
                                count: String(selectedSeasonEpisodeCount),
                            })}
                        </p>
                    </div>
                )}
                {isLoadingEpisodes ? (
                    <div className={styles.loadingState}>
                        <CircularProgress size={28} aria-label={t('LOADING')} />
                    </div>
                ) : (
                    <ListFeedback
                        isEmpty={episodes.length === 0}
                        isError={episodesQuery.isError}
                        isNextPageError={episodesQuery.isFetchNextPageError}
                        errorTitle={t('VIDEO_SECTION_LOAD_ERROR', { section: playlist.name })}
                        errorMessage={extractBackendErrorMessage(episodesQuery.error)}
                        onRetry={retryEpisodes}
                        hasNextPage={episodesQuery.hasNextPage}
                        isFetchingNextPage={episodesQuery.isFetchingNextPage}
                        onLoadMore={() => void episodesQuery.fetchNextPage()}
                        emptyState={null}
                    >
                        <div className={styles.detailListStack}>
                            {episodes.map((episode) => (
                                <VideoEpisodeRow
                                    key={episode.id}
                                    episode={episode}
                                    onOpenVideo={onOpenVideo}
                                    onSetWatched={onSetWatched}
                                />
                            ))}
                        </div>
                    </ListFeedback>
                )}
            </section>
        </div>
    );
}
