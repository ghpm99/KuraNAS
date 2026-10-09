import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import type { ReactNode } from 'react';
import { PlaylistsProvider, usePlaylistsProvider } from './playlistsProvider';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}));

const mockEnqueueSnackbar = jest.fn();

jest.mock('notistack', () => ({
    useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; put: jest.Mock };

const wrapper = ({ children }: { children: ReactNode }) => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return (
        <QueryClientProvider client={client}>
            <MemoryRouter>
                <PlaylistsProvider>{children}</PlaylistsProvider>
            </MemoryRouter>
        </QueryClientProvider>
    );
};

const renderWithSelectedPlaylist = async () => {
    const rendered = renderHook(() => usePlaylistsProvider(), { wrapper });
    await waitFor(() => expect(rendered.result.current.playlists.some((p) => p.id === 3)).toBe(true));
    act(() => rendered.result.current.selectPlaylist({ id: 3, name: 'Mix' } as never));
    await waitFor(() => expect(rendered.result.current.selectedPlaylist?.id).toBe(3));
    return rendered;
};

describe('features/music/playlistsProvider editing (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockImplementation((url: string) => {
            if (url === '/music/playlists/system') return Promise.resolve({ data: [] });
            if (url.includes('/tracks')) {
                return Promise.resolve({ data: { items: [], pagination: { has_next: false } } });
            }
            return Promise.resolve({
                data: {
                    items: [{ id: 3, name: 'Mix', is_system: false }],
                    pagination: { has_next: false },
                },
            });
        });
        mockedApi.put.mockResolvedValue({ data: { id: 3 } });
    });

    it('renameSelectedPlaylist PUTs name and description to /music/playlists/:id', async () => {
        const { result } = await renderWithSelectedPlaylist();
        const onSaved = jest.fn();

        act(() => result.current.renameSelectedPlaylist('Road trip', 'long drives', onSaved));

        await waitFor(() =>
            expect(mockedApi.put).toHaveBeenCalledWith('/music/playlists/3', {
                name: 'Road trip',
                description: 'long drives',
            })
        );
        await waitFor(() => expect(onSaved).toHaveBeenCalled());
        expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_PLAYLIST_UPDATED', {
            variant: 'success',
        });
    });

    it('renameSelectedPlaylist shows the backend message verbatim on failure', async () => {
        const { result } = await renderWithSelectedPlaylist();
        mockedApi.put.mockRejectedValue({ response: { data: { error: 'Requisição inválida' } } });

        act(() => result.current.renameSelectedPlaylist('', ''));

        await waitFor(() =>
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('Requisição inválida', {
                variant: 'error',
            })
        );
    });

    it('moveTrackToPosition PUTs one reorder item to /music/playlists/:id/tracks/reorder', async () => {
        const { result } = await renderWithSelectedPlaylist();

        act(() => result.current.moveTrackToPosition(55, 2));

        await waitFor(() =>
            expect(mockedApi.put).toHaveBeenCalledWith('/music/playlists/3/tracks/reorder', {
                tracks: [{ file_id: 55, position: 2 }],
            })
        );
    });

    it('moveTrackToPosition falls back to the translated failure and refreshes the tracks', async () => {
        const { result } = await renderWithSelectedPlaylist();
        mockedApi.put.mockRejectedValue(new Error('offline'));
        const trackFetchesBefore = mockedApi.get.mock.calls.filter(([url]) =>
            String(url).includes('/tracks')
        ).length;

        act(() => result.current.moveTrackToPosition(55, 2));

        await waitFor(() =>
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('MUSIC_PLAYLIST_REORDER_FAILED', {
                variant: 'error',
            })
        );
        await waitFor(() =>
            expect(
                mockedApi.get.mock.calls.filter(([url]) => String(url).includes('/tracks')).length
            ).toBeGreaterThan(trackFetchesBefore)
        );
    });
});
