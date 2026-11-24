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
        isRenamingPlaylist,
        createOpen,
        newName,
        newDescription,
        newKind,
        newRules,
        createErrorField,
        renameErrorField,
        selectPlaylist,
        backToList,
        fetchNextPlaylistPage,
        fetchNextTrackPage,
        openCreateDialog,
        closeCreateDialog,
        setNewName,
        setNewDescription,
        setNewKind,
        setNewRules,
        submitCreatePlaylist,
        deletePlaylistById,
        removeTrackByFileId,
        renameSelectedPlaylist,
        moveTrackToPosition,
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
                isRenaming={isRenamingPlaylist}
                renameErrorField={renameErrorField}
                onRenamePlaylist={renameSelectedPlaylist}
                onMoveTrack={moveTrackToPosition}
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
                newKind={newKind}
                newRules={newRules}
                errorField={createErrorField}
                isSubmitting={isCreatingPlaylist}
                onKindChange={setNewKind}
                onRulesChange={setNewRules}
                onClose={closeCreateDialog}
                onNameChange={setNewName}
                onDescriptionChange={setNewDescription}
                onSubmit={submitCreatePlaylist}
            />
        </>
    );
}
