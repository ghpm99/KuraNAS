import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import { apiBase } from '@/service';
import QueueDrawer from './QueueDrawer';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn() },
}));

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => ({
        queue: [
            { id: 11, queueEntryId: 'a', name: 'one' },
            { id: 12, queueEntryId: 'b', name: 'two' },
            { id: 11, queueEntryId: 'c', name: 'one again' },
        ],
        currentIndex: 0,
        queueOpen: true,
        setQueueOpen: jest.fn(),
        playTrackFromQueue: jest.fn(),
        removeFromQueue: jest.fn(),
        moveQueueItem: jest.fn(),
        clearQueue: jest.fn(),
        isPlaying: false,
    }),
}));

const mockedApi = apiBase as unknown as { post: jest.Mock };

describe('features/music/QueueDrawer save as playlist (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.post.mockImplementation((url: string) =>
            Promise.resolve({
                data: url === '/music/playlists/' ? { id: 9, name: 'Roadtrip' } : {},
            })
        );
    });

    it('creates the playlist then adds each distinct queued file in queue order', async () => {
        const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
        render(
            <QueryClientProvider client={client}>
                <SnackbarProvider>
                    <QueueDrawer />
                </SnackbarProvider>
            </QueryClientProvider>
        );

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_QUEUE_SAVE_AS_PLAYLIST' }));
        fireEvent.change(screen.getByLabelText('MUSIC_PLAYLIST_NAME'), {
            target: { value: ' Roadtrip ' },
        });
        fireEvent.click(
            screen.getByRole('button', { name: 'MUSIC_QUEUE_SAVE_AS_PLAYLIST_ACTION' })
        );

        await waitFor(() => expect(mockedApi.post).toHaveBeenCalledTimes(3));
        expect(mockedApi.post.mock.calls).toEqual([
            ['/music/playlists/', { name: 'Roadtrip' }],
            ['/music/playlists/9/tracks', { file_id: 11 }],
            ['/music/playlists/9/tracks', { file_id: 12 }],
        ]);
    });
});
