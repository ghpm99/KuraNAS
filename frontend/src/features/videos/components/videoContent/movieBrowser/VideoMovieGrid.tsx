import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import { getVideoMoviePosterUrl, type VideoFileDto } from '@/service/videoPlayback';
import VideoStarButton from '../../VideoStarButton';
import VideoMediaBadges from '../components/VideoMediaBadges';
import { useLocalPoster } from '../useLocalPoster';
import styles from './VideoMovieBrowser.module.css';

type VideoMovieGridProps = {
    movies: VideoFileDto[];
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    fetchNextPage: () => void;
    onPlayVideo: (videoId: number, playlistId?: number | null) => void;
};

const thumbnailBaseUrl = () => `${getApiV1BaseUrl()}/files/video-thumbnail`;

type VideoMovieCardProps = {
    movie: VideoFileDto;
    onPlayVideo: (videoId: number, playlistId?: number | null) => void;
};

function VideoMovieCard({ movie, onPlayVideo }: VideoMovieCardProps) {
    const { t } = useI18n();
    const { imageUrl, isPosterShown, handlePosterError } = useLocalPoster(
        getVideoMoviePosterUrl(movie.id),
        `${thumbnailBaseUrl()}/${movie.id}?width=320&height=180`
    );

    return (
        <div className={styles.movieCardWrapper}>
            <button
                type="button"
                className={styles.movieCard}
                aria-label={t('VIDEO_PLAY_ARIA', { title: movie.name })}
                onClick={() => onPlayVideo(movie.id, null)}
            >
                <div className={styles.movieThumb}>
                    <img
                        loading="lazy"
                        src={imageUrl}
                        alt={movie.name}
                        onError={isPosterShown ? handlePosterError : undefined}
                    />
                    <VideoMediaBadges video={movie} />
                </div>
                <h3 className={styles.movieName}>{movie.name}</h3>
            </button>
            <VideoStarButton
                videoId={movie.id}
                videoName={movie.name}
                isStarred={movie.starred}
                placement="overlay"
            />
        </div>
    );
}

export default function VideoMovieGrid({
    movies,
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
    onPlayVideo,
}: VideoMovieGridProps) {
    return (
        <>
            <div className={styles.movieGrid}>
                {movies.map((movie) => (
                    <VideoMovieCard key={movie.id} movie={movie} onPlayVideo={onPlayVideo} />
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
