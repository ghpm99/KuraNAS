import type { IMusicData } from '@/features/music/providers/musicProvider/musicProvider';
import {
    createAlbumPlaybackContext,
    createAllTracksPlaybackContext,
    type MusicPlaybackContext,
} from '@/features/music/components/playbackContext';
import type { GlobalSearchTrackResult } from '@/service/search';

const extractFormat = (path: string) => {
    const extensionStart = path.lastIndexOf('.');
    return extensionStart < 0 ? '' : path.slice(extensionStart).toLowerCase();
};

const extractFileName = (path: string) => path.slice(path.lastIndexOf('/') + 1);

export const buildPlayableTrack = (track: GlobalSearchTrackResult): IMusicData => ({
    id: track.file_id,
    name: extractFileName(track.path) || track.title,
    path: track.path,
    type: 2,
    format: extractFormat(track.path),
    size: 0,
    updated_at: '',
    created_at: '',
    deleted_at: '',
    last_interaction: '',
    last_backup: '',
    check_sum: '',
    directory_content_count: 0,
    starred: false,
    metadata: {
        title: track.title,
        artist: track.artist,
        album: track.album,
        length: track.duration,
    },
});

export const buildTrackPlaybackContext = (track: GlobalSearchTrackResult): MusicPlaybackContext =>
    track.album ? createAlbumPlaybackContext(track.album) : createAllTracksPlaybackContext();

export const formatTrackDuration = (durationSeconds: number) => {
    const totalSeconds = Math.max(0, Math.round(durationSeconds || 0));
    if (totalSeconds === 0) {
        return '';
    }
    const seconds = String(totalSeconds % 60).padStart(2, '0');
    return `${Math.floor(totalSeconds / 60)}:${seconds}`;
};
