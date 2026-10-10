import { TextField } from '@mui/material';
import { ArrowDown, ArrowUp, Play } from 'lucide-react';
import {
    VideoFileDto,
    VideoLibrarySort,
    VideoLibrarySortKey,
    VideoPlaylistDto,
} from '@/service/videoPlayback';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';
import VideoLibraryPlaylistPicker from './VideoLibraryPlaylistPicker';
import VideoWatchedToggleButton from './VideoWatchedToggleButton';
import {
    defaultVideoLibrarySort,
    videoLibrarySortKeys,
} from '../../../providers/videoContentProvider/videoLibrarySortPreference';
import styles from '../videoContent.module.css';

type VideoLibrarySectionProps = {
    videos: VideoFileDto[];
    playlists: VideoPlaylistDto[];
    search: string;
    sort?: VideoLibrarySort;
    selectedPlaylistPerVideo: Record<number, number>;
    isAddingToPlaylist: boolean;
    isFetchingMoreVideos: boolean;
    hasMoreVideos: boolean;
    onSearchChange: (value: string) => void;
    onSortChange?: (sort: VideoLibrarySort) => void;
    onSelectPlaylistForVideo: (videoId: number, playlistId: number) => void;
    onPlayVideo: (videoId: number, playlistId?: number | null) => void;
    onAddVideo: (videoId: number) => void;
    onLoadMore: () => void;
    onSetWatched?: (videoId: number, watched: boolean) => void;
};

const apiBase = `${getApiV1BaseUrl()}/files`;

const sortLabelKeys: Record<VideoLibrarySortKey, string> = {
    recent: 'VIDEO_SORT_RECENT',
    name: 'VIDEO_SORT_NAME',
    size: 'VIDEO_SORT_SIZE',
    duration: 'VIDEO_SORT_DURATION',
};

export default function VideoLibrarySection({
    videos,
    playlists,
    search,
    sort = defaultVideoLibrarySort,
    selectedPlaylistPerVideo,
    isAddingToPlaylist,
    isFetchingMoreVideos,
    hasMoreVideos,
    onSearchChange,
    onSortChange,
    onSelectPlaylistForVideo,
    onPlayVideo,
    onAddVideo,
    onLoadMore,
    onSetWatched,
}: VideoLibrarySectionProps) {
    const { t } = useI18n();
    const isAscending = sort.order === 'asc';
    const orderToggleLabel = t(
        isAscending ? 'VIDEO_SORT_ORDER_ASCENDING' : 'VIDEO_SORT_ORDER_DESCENDING'
    );

    return (
        <section className={styles.sectionBlock}>
            <div className={styles.sectionHeader}>
                <h2>{t('VIDEO_ALL')}</h2>
                <p>{t('VIDEO_ALL_DESC')}</p>
            </div>
            <div className={styles.sortRow}>
                <select
                    className={styles.playlistSelect}
                    aria-label={t('VIDEO_SORT_LABEL')}
                    value={sort.key}
                    onChange={(event) =>
                        onSortChange?.({
                            key: event.target.value as VideoLibrarySortKey,
                            order: sort.order,
                        })
                    }
                >
                    {videoLibrarySortKeys.map((sortKey) => (
                        <option key={sortKey} value={sortKey}>
                            {t(sortLabelKeys[sortKey])}
                        </option>
                    ))}
                </select>
                <button
                    type="button"
                    className={styles.actionBtn}
                    aria-label={orderToggleLabel}
                    title={orderToggleLabel}
                    onClick={() =>
                        onSortChange?.({ key: sort.key, order: isAscending ? 'desc' : 'asc' })
                    }
                >
                    {isAscending ? <ArrowUp size={14} /> : <ArrowDown size={14} />}
                </button>
            </div>
            <div className={styles.searchRow}>
                <TextField
                    size="small"
                    fullWidth
                    placeholder={t('VIDEO_SEARCH_PLACEHOLDER')}
                    value={search}
                    onChange={(event) => onSearchChange(event.target.value)}
                />
            </div>
            <div className={styles.allVideosList}>
                {videos.map((video) => {
                    return (
                        <div className={styles.allVideoItem} key={video.id}>
                            <div className={styles.allVideoThumb}>
                                <img
                                    loading="lazy"
                                    src={`${apiBase}/video-thumbnail/${video.id}?width=240&height=135`}
                                    alt={video.name}
                                />
                            </div>
                            <div className={styles.allVideoMeta}>
                                <h4>{video.name}</h4>
                                <p>
                                    {video.parent_path} · {video.format.toUpperCase()}
                                </p>
                            </div>
                            <div className={styles.allVideoActions}>
                                <button
                                    type="button"
                                    className={styles.actionBtn}
                                    onClick={() => onPlayVideo(video.id, null)}
                                >
                                    <Play size={14} />
                                    {t('VIDEO_PLAY')}
                                </button>
                                <VideoLibraryPlaylistPicker
                                    videoId={video.id}
                                    playlists={playlists}
                                    selectedPlaylistId={selectedPlaylistPerVideo[video.id]}
                                    isAddingToPlaylist={isAddingToPlaylist}
                                    onSelectPlaylist={onSelectPlaylistForVideo}
                                    onAddVideo={onAddVideo}
                                />
                                {onSetWatched && (
                                    <VideoWatchedToggleButton
                                        isWatched={false}
                                        onToggle={() => onSetWatched(video.id, true)}
                                    />
                                )}
                            </div>
                        </div>
                    );
                })}
            </div>
            <LoadMoreSentinel
                hasNextPage={hasMoreVideos}
                isFetchingNextPage={isFetchingMoreVideos}
                fetchNextPage={onLoadMore}
            />
        </section>
    );
}
