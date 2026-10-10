import { useState } from 'react';
import { Plus } from 'lucide-react';
import type { VideoPlaylistDto } from '@/service/videoPlayback';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useVideoPlaylistsOfVideo } from '../../../providers/videoContentProvider/useVideoQueries';
import styles from '../videoContent.module.css';

type VideoLibraryPlaylistPickerProps = {
    videoId: number;
    playlists: VideoPlaylistDto[];
    selectedPlaylistId?: number;
    isAddingToPlaylist: boolean;
    onSelectPlaylist: (videoId: number, playlistId: number) => void;
    onAddVideo: (videoId: number) => void;
};

export default function VideoLibraryPlaylistPicker({
    videoId,
    playlists,
    selectedPlaylistId,
    isAddingToPlaylist,
    onSelectPlaylist,
    onAddVideo,
}: VideoLibraryPlaylistPickerProps) {
    const { t } = useI18n();
    const [hasOpenedPicker, setHasOpenedPicker] = useState(false);
    const { data: playlistsOfVideo = [] } = useVideoPlaylistsOfVideo(videoId, hasOpenedPicker);

    const selectedPlaylist = selectedPlaylistId ?? playlists[0]?.id;
    const isAlreadyInPlaylist = playlistsOfVideo.some(
        (playlistOfVideo) => playlistOfVideo.id === selectedPlaylist
    );

    return (
        <>
            <select
                className={styles.playlistSelect}
                value={selectedPlaylist ?? ''}
                onFocus={() => setHasOpenedPicker(true)}
                onMouseDown={() => setHasOpenedPicker(true)}
                onChange={(event) => onSelectPlaylist(videoId, Number(event.target.value))}
            >
                {playlists.map((playlist) => (
                    <option key={`add-${videoId}-${playlist.id}`} value={playlist.id}>
                        {playlist.name}
                    </option>
                ))}
            </select>
            <button
                type="button"
                className={styles.actionBtn}
                disabled={!selectedPlaylist || isAddingToPlaylist || isAlreadyInPlaylist}
                onClick={() => onAddVideo(videoId)}
            >
                <Plus size={14} />
                {isAlreadyInPlaylist ? t('VIDEO_ALREADY_ADDED') : t('VIDEO_ADD')}
            </button>
        </>
    );
}
