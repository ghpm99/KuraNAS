import { Pagination } from '@/types/pagination';
import { Playlist, PlaylistTrack } from '@/types/playlist';

export interface PlaylistsContextData {
    selectedPlaylist: Playlist | null;
    playlists: Playlist[];
    tracks: PlaylistTrack[];
    isLoadingPlaylists: boolean;
    isLoadingTracks: boolean;
    hasNextPlaylistPage: boolean;
    hasNextTrackPage: boolean;
    isFetchingNextPlaylistPage: boolean;
    isFetchingNextTrackPage: boolean;
    isCreatingPlaylist: boolean;
    isDeletingPlaylist: boolean;
    isRemovingTrack: boolean;
    createOpen: boolean;
    newName: string;
    newDescription: string;
    selectPlaylist: (playlist: Playlist) => void;
    backToList: () => void;
    fetchNextPlaylistPage: () => Promise<unknown>;
    fetchNextTrackPage: () => Promise<unknown>;
    openCreateDialog: () => void;
    closeCreateDialog: () => void;
    setNewName: (value: string) => void;
    setNewDescription: (value: string) => void;
    submitCreatePlaylist: () => void;
    deletePlaylistById: (id: number) => void;
    removeTrackByFileId: (fileId: number) => void;
    playlistQueryFn: (pageParam: number) => Promise<Pagination<Playlist>>;
    playlistTracksQueryFn: (pageParam: number) => Promise<Pagination<PlaylistTrack>>;
}
