import { Film } from 'lucide-react';
import EmptyState from '@/components/emptyState/emptyState';
import ErrorState from '@/components/errorState/errorState';
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
    const { sort, setSort, movieQuery, movies } = useVideoMovieBrowser();

    const renderMovies = () => {
        if (movieQuery.status === 'pending') {
            return <VideoSectionSkeleton cardCount={6} />;
        }
        if (movieQuery.status === 'error') {
            return (
                <ErrorState
                    title={t('VIDEO_SECTION_LOAD_ERROR', { section: t('VIDEO_SECTION_MOVIES') })}
                    backendMessage={extractBackendErrorMessage(movieQuery.error)}
                    onRetry={() => void movieQuery.refetch()}
                />
            );
        }
        if (movies.length === 0) {
            return (
                <EmptyState
                    icon={<Film size={28} aria-hidden="true" />}
                    title={t('VIDEO_SECTION_MOVIES_EMPTY')}
                />
            );
        }
        return (
            <VideoMovieGrid
                movies={movies}
                hasNextPage={movieQuery.hasNextPage}
                isFetchingNextPage={movieQuery.isFetchingNextPage}
                fetchNextPage={() => void movieQuery.fetchNextPage()}
                onPlayVideo={onPlayVideo}
            />
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
