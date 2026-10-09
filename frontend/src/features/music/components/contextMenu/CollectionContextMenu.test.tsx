import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { SnackbarProvider } from 'notistack';
import type { IMusicData } from '@/types/music';
import CollectionContextMenu from './CollectionContextMenu';

const mockPlayNext = jest.fn();
const mockAddToQueue = jest.fn();

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => ({ playNext: mockPlayNext, addToQueue: mockAddToQueue }),
}));

const playbackContext = {
    kind: 'album' as const,
    labelKey: 'MUSIC_PLAYBACK_CONTEXT_ALBUM',
    href: '/music/albums',
};
const tracks = [{ id: 1 }, { id: 2 }] as IMusicData[];

const renderMenu = (loadTracks: () => Promise<IMusicData[]>) =>
    render(
        <SnackbarProvider>
            <CollectionContextMenu
                collectionName="Hits"
                playbackContext={playbackContext}
                loadTracks={loadTracks}
                layout="card"
            >
                <div>card</div>
            </CollectionContextMenu>
        </SnackbarProvider>
    );

describe('CollectionContextMenu', () => {
    beforeEach(() => jest.clearAllMocks());

    it('opens from the more actions button and queues the full collection next', async () => {
        renderMenu(() => Promise.resolve(tracks));

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_COLLECTION_MORE_ACTIONS' }));
        fireEvent.click(screen.getByRole('menuitem', { name: 'MUSIC_PLAY_NEXT' }));

        await waitFor(() => expect(mockPlayNext).toHaveBeenCalledWith(tracks, playbackContext));
    });

    it('opens on right click and appends the collection to the queue', async () => {
        renderMenu(() => Promise.resolve(tracks));

        fireEvent.contextMenu(screen.getByText('card'));
        fireEvent.click(screen.getByRole('menuitem', { name: 'MUSIC_ADD_TO_QUEUE' }));

        await waitFor(() => expect(mockAddToQueue).toHaveBeenCalledWith(tracks, playbackContext));
    });

    it('does not queue anything when loading the collection fails', async () => {
        const loadTracks = jest.fn().mockRejectedValue(new Error('offline'));
        renderMenu(loadTracks);

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_COLLECTION_MORE_ACTIONS' }));
        fireEvent.click(screen.getByRole('menuitem', { name: 'MUSIC_ADD_TO_QUEUE' }));

        await waitFor(() => expect(loadTracks).toHaveBeenCalled());
        expect(mockAddToQueue).not.toHaveBeenCalled();
    });
});
