import { TextField } from '@mui/material';
import { Play, Plus } from 'lucide-react';
import { VideoFileDto, VideoPlaylistDto } from '@/service/videoPlayback';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import styles from '../videoContent.module.css';

type VideoLibrarySectionProps = {
    videos: VideoFileDto[];
    playlists: VideoPlaylistDto[];
    playlistMembershipMap: Record<number, Set<number>>;
    search: string;
    selectedPlaylistPerVideo: Record<number, number>;
    isAddingToPlaylist: boolean;
    isFetchingMoreVideos: boolean;
    hasMoreVideos: boolean;
    onSearchChange: (value: string) => void;
    onSelectPlaylistForVideo: (videoId: number, playlistId: number) => void;
    onPlayVideo: (videoId: number, playlistId?: number | null) => void;
    onAddVideo: (videoId: number) => void;
    onLoadMore: () => void;
};

const apiBase = `${getApiV1BaseUrl()}/files`;

export default function VideoLibrarySection({
    videos,
    playlists,
    playlistMembershipMap,
    search,
    selectedPlaylistPerVideo,
    isAddingToPlaylist,
    isFetchingMoreVideos,
    hasMoreVideos,
    onSearchChange,
    onSelectPlaylistForVideo,
    onPlayVideo,
    onAddVideo,
    onLoadMore,
}: VideoLibrarySectionProps) {
    const { t } = useI18n();

    return (
        <section className={styles.sectionBlock}>
            <div className={styles.sectionHeader}>
                <h2>{t('VIDEO_ALL')}</h2>
                <p>{t('VIDEO_ALL_DESC')}</p>
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
                    const selectedPlaylist = selectedPlaylistPerVideo[video.id] ?? playlists[0]?.id;
                    const isAlreadyInPlaylist = Boolean(
                        selectedPlaylist &&
                        playlistMembershipMap[selectedPlaylist] &&
                        playlistMembershipMap[selectedPlaylist]?.has(video.id)
                    );
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
                                <select
                                    className={styles.playlistSelect}
                                    value={selectedPlaylist ?? ''}
                                    onChange={(event) =>
                                        onSelectPlaylistForVideo(
                                            video.id,
                                            Number(event.target.value)
                                        )
                                    }
                                >
                                    {playlists.map((playlist) => (
                                        <option
                                            key={`add-${video.id}-${playlist.id}`}
                                            value={playlist.id}
                                        >
                                            {playlist.name}
                                        </option>
                                    ))}
                                </select>
                                <button
                                    type="button"
                                    className={styles.actionBtn}
                                    disabled={
                                        !selectedPlaylist ||
                                        isAddingToPlaylist ||
                                        isAlreadyInPlaylist
                                    }
                                    onClick={() => onAddVideo(video.id)}
                                >
                                    <Plus size={14} />
                                    {isAlreadyInPlaylist
                                        ? t('VIDEO_ALREADY_ADDED')
                                        : t('VIDEO_ADD')}
                                </button>
                            </div>
                        </div>
                    );
                })}
            </div>
            {hasMoreVideos ? (
                <div className={styles.libraryFooter}>
                    <button
                        type="button"
                        className={styles.actionBtn}
                        onClick={onLoadMore}
                        disabled={isFetchingMoreVideos}
                    >
                        {isFetchingMoreVideos ? t('LOADING') : t('ACTION_LOAD_MORE')}
                    </button>
                </div>
            ) : null}
        </section>
    );
}
