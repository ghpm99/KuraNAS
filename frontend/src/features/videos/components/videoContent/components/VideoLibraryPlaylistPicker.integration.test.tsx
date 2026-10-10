import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import VideoLibraryPlaylistPicker from './VideoLibraryPlaylistPicker';
import { apiBase } from '@/service';
import type { VideoPlaylistDto } from '@/service/videoPlayback';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock };

const playlists = [
    { id: 1, name: 'Season 1', type: 'series' },
    { id: 2, name: 'Favorites', type: 'custom' },
] as VideoPlaylistDto[];

const renderPicker = (selectedPlaylistId?: number) =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <VideoLibraryPlaylistPicker
                videoId={5}
                playlists={playlists}
                selectedPlaylistId={selectedPlaylistId}
                isAddingToPlaylist={false}
                onSelectPlaylist={jest.fn()}
                onAddVideo={jest.fn()}
            />
        </QueryClientProvider>
    );

describe('features/videos/VideoLibraryPlaylistPicker (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockResolvedValue({ data: [{ id: 2, name: 'Favorites', type: 'custom' }] });
    });

    it('does not ask the backend until the playlist select is opened', () => {
        renderPicker();

        expect(mockedApi.get).not.toHaveBeenCalled();
    });

    it('fetches the playlists of the video on open and disables adding when already in the selected playlist', async () => {
        renderPicker(2);

        fireEvent.mouseDown(screen.getByRole('combobox'));

        await waitFor(() =>
            expect(screen.getByRole('button', { name: 'VIDEO_ALREADY_ADDED' })).toBeDisabled()
        );
        expect(mockedApi.get).toHaveBeenCalledTimes(1);
        expect(mockedApi.get).toHaveBeenCalledWith('/video/playlists/by-video/5');
    });

    it('keeps adding enabled when the video is not in the selected playlist', async () => {
        renderPicker(1);

        fireEvent.focus(screen.getByRole('combobox'));

        await waitFor(() =>
            expect(mockedApi.get).toHaveBeenCalledWith('/video/playlists/by-video/5')
        );
        expect(screen.getByRole('button', { name: 'VIDEO_ADD' })).toBeEnabled();
    });
});
