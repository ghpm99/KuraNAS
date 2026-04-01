import { usePlaylistsProvider } from '@/features/music/providers/playlistsProvider';
import PlaylistCreateDialog from './PlaylistCreateDialog';
import PlaylistDetailSection from './PlaylistDetailSection';
import PlaylistListSection from './PlaylistListSection';

export default function PlaylistsScreen() {
    const {
        selectedPlaylist,
        playlists,
        tracks,
        isLoadingPlaylists,
        isLoadingTracks,
        hasNextPlaylistPage,
        hasNextTrackPage,
        isFetchingNextPlaylistPage,
        isFetchingNextTrackPage,
        isCreatingPlaylist,
        createOpen,
        newName,
        newDescription,
        selectPlaylist,
        backToList,
        fetchNextPlaylistPage,
        fetchNextTrackPage,
        openCreateDialog,
        closeCreateDialog,
        setNewName,
        setNewDescription,
        submitCreatePlaylist,
        deletePlaylistById,
        removeTrackByFileId,
    } = usePlaylistsProvider();

    if (selectedPlaylist) {
        return (
            <PlaylistDetailSection
                playlist={selectedPlaylist}
                tracks={tracks}
                isLoading={isLoadingTracks}
                hasNextPage={hasNextTrackPage}
                isFetchingNextPage={isFetchingNextTrackPage}
                onBack={backToList}
                onRemoveTrack={removeTrackByFileId}
                onLoadMore={() => void fetchNextTrackPage()}
            />
        );
    }

    return (
        <>
            <PlaylistListSection
                playlists={playlists}
                isLoading={isLoadingPlaylists}
                hasNextPage={hasNextPlaylistPage}
                isFetchingNextPage={isFetchingNextPlaylistPage}
                onSelect={selectPlaylist}
                onDelete={deletePlaylistById}
                onLoadMore={() => void fetchNextPlaylistPage()}
                onCreateOpen={openCreateDialog}
            />
            <PlaylistCreateDialog
                open={createOpen}
                newName={newName}
                newDescription={newDescription}
                isSubmitting={isCreatingPlaylist}
                onClose={closeCreateDialog}
                onNameChange={setNewName}
                onDescriptionChange={setNewDescription}
                onSubmit={submitCreatePlaylist}
            />
        </>
    );
}
