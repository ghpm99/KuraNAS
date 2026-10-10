import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import type { VideoFileDto } from '@/service/videoPlayback';
import styles from './VideoMovieBrowser.module.css';

type VideoMovieGridProps = {
    movies: VideoFileDto[];
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    fetchNextPage: () => void;
    onPlayVideo: (videoId: number, playlistId?: number | null) => void;
};

const thumbnailBaseUrl = () => `${getApiV1BaseUrl()}/files/video-thumbnail`;

export default function VideoMovieGrid({
    movies,
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
    onPlayVideo,
}: VideoMovieGridProps) {
    const { t } = useI18n();

    return (
        <>
            <div className={styles.movieGrid}>
                {movies.map((movie) => (
                    <button
                        type="button"
                        key={movie.id}
                        className={styles.movieCard}
                        aria-label={t('VIDEO_PLAY_ARIA', { title: movie.name })}
                        onClick={() => onPlayVideo(movie.id, null)}
                    >
                        <div className={styles.movieThumb}>
                            <img
                                loading="lazy"
                                src={`${thumbnailBaseUrl()}/${movie.id}?width=320&height=180`}
                                alt={movie.name}
                            />
                        </div>
                        <h3 className={styles.movieName}>{movie.name}</h3>
                    </button>
                ))}
            </div>
            <LoadMoreSentinel
                hasNextPage={hasNextPage}
                isFetchingNextPage={isFetchingNextPage}
                fetchNextPage={fetchNextPage}
            />
        </>
    );
}
