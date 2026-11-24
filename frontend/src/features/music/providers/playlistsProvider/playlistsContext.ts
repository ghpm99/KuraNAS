import { Pagination } from '@/types/pagination';
import { Playlist, PlaylistRules, PlaylistTrack } from '@/types/playlist';

export type PlaylistKindChoice = 'manual' | 'rule';

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
    isRenamingPlaylist: boolean;
    isMovingTrack: boolean;
    createOpen: boolean;
    newName: string;
    newDescription: string;
    newKind: PlaylistKindChoice;
    newRules: PlaylistRules;
    createErrorField: string | null;
    renameErrorField: string | null;
    selectPlaylist: (playlist: Playlist) => void;
    backToList: () => void;
    fetchNextPlaylistPage: () => Promise<unknown>;
    fetchNextTrackPage: () => Promise<unknown>;
    openCreateDialog: () => void;
    openCreateRuleDialog: (initialRules?: Partial<PlaylistRules>) => void;
    closeCreateDialog: () => void;
    setNewKind: (kind: PlaylistKindChoice) => void;
    setNewRules: (rules: PlaylistRules) => void;
    setNewName: (value: string) => void;
    setNewDescription: (value: string) => void;
    submitCreatePlaylist: () => void;
    deletePlaylistById: (id: number) => void;
    removeTrackByFileId: (fileId: number) => void;
    renameSelectedPlaylist: (
        name: string,
        description: string,
        onSaved?: () => void,
        rules?: PlaylistRules
    ) => void;
    moveTrackToPosition: (fileId: number, position: number) => void;
    playlistQueryFn: (pageParam: number) => Promise<Pagination<Playlist>>;
    playlistTracksQueryFn: (pageParam: number) => Promise<Pagination<PlaylistTrack>>;
}
