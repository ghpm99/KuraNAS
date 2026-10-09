import { addTrackToPlaylist, createPlaylist } from '@/service/playlist';
import { runWithConcurrencyLimit } from '@/shared/utils/runWithConcurrencyLimit';

const ORDER_PRESERVING_CONCURRENCY = 1;

export type SavedPlaylistSummary = {
    playlistName: string;
    totalTracks: number;
    addedTracks: number;
};

export const saveTracksAsPlaylist = async (
    playlistName: string,
    fileIds: number[]
): Promise<SavedPlaylistSummary> => {
    const uniqueFileIds = [...new Set(fileIds)];
    const playlist = await createPlaylist({ name: playlistName });
    const outcomes = await runWithConcurrencyLimit(
        uniqueFileIds,
        ORDER_PRESERVING_CONCURRENCY,
        (fileId) => addTrackToPlaylist(playlist.id, fileId)
    );
    return {
        playlistName: playlist.name ?? playlistName,
        totalTracks: uniqueFileIds.length,
        addedTracks: outcomes.filter((outcome) => outcome.status === 'fulfilled').length,
    };
};
