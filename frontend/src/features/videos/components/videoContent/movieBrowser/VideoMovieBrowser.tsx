import { Film } from 'lucide-react';
import EmptyState from '@/components/emptyState/emptyState';
import ListFeedback from '@/components/listFeedback/ListFeedback';
import useI18n from '@/components/i18n/provider/i18nContext';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import VideoSectionSkeleton from '../components/VideoSectionSkeleton';
import VideoMovieGrid from './VideoMovieGrid';
import VideoMovieSortSelect from './VideoMovieSortSelect';
import { useVideoMovieBrowser } from './useVideoMovieBrowser';
import styles from './VideoMovieBrowser.module.css';

type VideoMovieBrowserProps = {
    onPlayVideo: (videoId: number, playlistId?: number | null) => void;
};

export default function VideoMovieBrowser({ onPlayVideo }: VideoMovieBrowserProps) {
    const { t } = useI18n();
    const { sort, setSort, movieQuery, movies, retryFailedPage } = useVideoMovieBrowser();

    const renderMovies = () => {
        if (movieQuery.status === 'pending') {
            return <VideoSectionSkeleton cardCount={6} />;
        }
        return (
            <ListFeedback
                isEmpty={movies.length === 0}
                isError={movieQuery.status === 'error'}
                isNextPageError={movieQuery.isFetchNextPageError}
                errorTitle={t('VIDEO_SECTION_LOAD_ERROR', { section: t('VIDEO_SECTION_MOVIES') })}
                errorMessage={extractBackendErrorMessage(movieQuery.error)}
                onRetry={retryFailedPage}
                hasNextPage={movieQuery.hasNextPage}
                isFetchingNextPage={movieQuery.isFetchingNextPage}
                onLoadMore={() => void movieQuery.fetchNextPage()}
                emptyState={
                    <EmptyState
                        icon={<Film size={28} aria-hidden="true" />}
                        title={t('VIDEO_SECTION_MOVIES_EMPTY')}
                    />
                }
            >
                <VideoMovieGrid movies={movies} onPlayVideo={onPlayVideo} />
            </ListFeedback>
        );
    };

    return (
        <section className={styles.browser}>
            <header className={styles.header}>
                <div>
                    <h2>{t('VIDEO_SECTION_MOVIES')}</h2>
                    <p>{t('VIDEO_SECTION_MOVIES_DESCRIPTION')}</p>
                </div>
                <VideoMovieSortSelect sort={sort} onChange={setSort} />
            </header>
            {renderMovies()}
        </section>
    );
}
