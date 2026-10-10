import { Play } from 'lucide-react';
import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import type { VideoFileDto } from '@/service/videoPlayback';
import styles from './VideoFolderBrowser.module.css';

type VideoFolderVideoListProps = {
    videos: VideoFileDto[];
    hasNextPage: boolean;
    isFetchingNextPage: boolean;
    fetchNextPage: () => void;
    onPlayVideo: (videoId: number, playlistId?: number | null) => void;
};

const thumbnailBaseUrl = () => `${getApiV1BaseUrl()}/files/video-thumbnail`;

export default function VideoFolderVideoList({
    videos,
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
    onPlayVideo,
}: VideoFolderVideoListProps) {
    const { t } = useI18n();

    return (
        <>
            <div className={styles.videoList}>
                {videos.map((video) => (
                    <button
                        type="button"
                        key={video.id}
                        className={styles.videoRow}
                        aria-label={t('VIDEO_PLAY_ARIA', { title: video.name })}
                        onClick={() => onPlayVideo(video.id, null)}
                    >
                        <div className={styles.videoThumb}>
                            <img
                                loading="lazy"
                                src={`${thumbnailBaseUrl()}/${video.id}?width=240&height=135`}
                                alt={video.name}
                            />
                        </div>
                        <div className={styles.videoMeta}>
                            <h4>{video.name}</h4>
                            <p>{video.format.toUpperCase()}</p>
                        </div>
                        <Play size={16} aria-hidden="true" />
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
