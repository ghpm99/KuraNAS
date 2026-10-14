import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import type { ReactNode } from 'react';
import { VideoContentProvider, useVideoContentProvider } from './videoContentProvider';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}));

const mockNavigate = jest.fn();

jest.mock('react-router-dom', () => {
    const actual = jest.requireActual('react-router-dom');
    return { ...actual, useNavigate: () => mockNavigate };
});

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as {
    get: jest.Mock;
    put: jest.Mock;
};

const buildPlaylist = (id: number, name: string, isHidden: boolean) => ({
    id,
    type: 'series',
    source_path: `/videos/${name}`,
    name,
    is_hidden: isHidden,
    is_auto: true,
    group_mode: 'folder',
    classification: 'series',
    item_count: 2,
    cover_video_id: 10,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    last_played_at: null,
    items: [],
});

const pageOf = (items: unknown[], hasNext = false) => ({
    data: { items, pagination: { page: 1, page_size: 24, has_next: hasNext, has_prev: false } },
});

const renderProviderAt = (route: string) =>
    renderHook(() => useVideoContentProvider(), {
        wrapper: ({ children }: { children: ReactNode }) => (
            <QueryClientProvider
                client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
            >
                <MemoryRouter initialEntries={[route]}>
                    <VideoContentProvider>{children}</VideoContentProvider>
                </MemoryRouter>
            </QueryClientProvider>
        ),
    });

describe('features/videos/videoContentProvider playlist visibility (seam)', () => {
    let hiddenOnServer: ReturnType<typeof buildPlaylist>[];

    beforeEach(() => {
        jest.clearAllMocks();
        hiddenOnServer = [buildPlaylist(8, 'oculta', true)];
        mockedApi.put.mockResolvedValue({ data: undefined });
        mockedApi.get.mockImplementation((url: string) => {
            if (url === '/video/playlists/hidden') return Promise.resolve(pageOf(hiddenOnServer));
            if (url === '/video/playlists/') {
                return Promise.resolve({ data: [buildPlaylist(2, 'serie', false)] });
            }
            if (url === '/video/playlists/section/series') {
                return Promise.resolve(pageOf([buildPlaylist(2, 'serie', false)]));
            }
            if (url === '/video/playlists/2/items') return Promise.resolve(pageOf([]));
            if (url === '/video/continue') return Promise.resolve({ data: [] });
            if (url === '/video/catalog/home') return Promise.resolve({ data: { sections: [] } });
            return Promise.resolve(pageOf([]));
        });
    });

    it('loads the hidden playlists only on the hidden route', async () => {
        const { result } = renderProviderAt('/videos/hidden');

        await waitFor(() => expect(result.current.hiddenPlaylists.map((p) => p.id)).toEqual([8]));
        expect(mockedApi.get).toHaveBeenCalledWith('/video/playlists/hidden', {
            params: { page: 1, page_size: 24 },
        });
    });

    it('does not request the hidden playlists outside the hidden route', async () => {
        const { result } = renderProviderAt('/videos/series');

        await waitFor(() => expect(result.current.seriesPlaylists).toHaveLength(1));
        expect(mockedApi.get).not.toHaveBeenCalledWith(
            '/video/playlists/hidden',
            expect.anything()
        );
    });

    it('hidePlaylist PUTs hidden true, refreshes the lists and offers undo that PUTs hidden false', async () => {
        const { result } = renderProviderAt('/videos/series');
        await waitFor(() => expect(result.current.seriesPlaylists).toHaveLength(1));
        const sectionCallsBefore = mockedApi.get.mock.calls.filter(
            ([url]) => url === '/video/playlists/section/series'
        ).length;

        act(() => result.current.hidePlaylist(result.current.seriesPlaylists[0]!));

        await waitFor(() =>
            expect(mockedApi.put).toHaveBeenCalledWith('/video/playlists/2/hidden', {
                hidden: true,
            })
        );
        await waitFor(() => expect(result.current.feedback.message).toBe('VIDEO_PLAYLIST_HIDDEN'));
        expect(result.current.feedback.action?.label).toBe('VIDEO_UNDO');
        await waitFor(() =>
            expect(
                mockedApi.get.mock.calls.filter(
                    ([url]) => url === '/video/playlists/section/series'
                ).length
            ).toBeGreaterThan(sectionCallsBefore)
        );

        act(() => result.current.feedback.action?.onClick());

        await waitFor(() =>
            expect(mockedApi.put).toHaveBeenCalledWith('/video/playlists/2/hidden', {
                hidden: false,
            })
        );
        await waitFor(() => expect(result.current.feedback.message).toBe('VIDEO_PLAYLIST_SHOWN'));
        expect(result.current.feedback.action).toBeUndefined();
    });

    it('hideSelectedPlaylist hides the open playlist and goes back to its section', async () => {
        const { result } = renderProviderAt('/videos/series/serie?playlist=2');
        await waitFor(() => expect(result.current.selectedPlaylistSummary?.id).toBe(2));

        act(() => result.current.hideSelectedPlaylist());

        await waitFor(() =>
            expect(mockedApi.put).toHaveBeenCalledWith('/video/playlists/2/hidden', {
                hidden: true,
            })
        );
        await waitFor(() => expect(mockNavigate).toHaveBeenCalledWith('/videos/series'));
    });

    it('showPlaylist PUTs hidden false and drops the playlist from the hidden list', async () => {
        const { result } = renderProviderAt('/videos/hidden');
        await waitFor(() => expect(result.current.hiddenPlaylists).toHaveLength(1));
        mockedApi.put.mockImplementation(() => {
            hiddenOnServer = [];
            return Promise.resolve({ data: undefined });
        });

        act(() => result.current.showPlaylist(result.current.hiddenPlaylists[0]!));

        await waitFor(() =>
            expect(mockedApi.put).toHaveBeenCalledWith('/video/playlists/8/hidden', {
                hidden: false,
            })
        );
        await waitFor(() => expect(result.current.hiddenPlaylists).toHaveLength(0));
        expect(result.current.feedback.message).toBe('VIDEO_PLAYLIST_SHOWN');
    });

    it('shows the backend message verbatim when changing visibility fails', async () => {
        const { result } = renderProviderAt('/videos/hidden');
        await waitFor(() => expect(result.current.hiddenPlaylists).toHaveLength(1));
        mockedApi.put.mockRejectedValue({ response: { data: { error: 'Nao encontrado' } } });

        act(() => result.current.showPlaylist(result.current.hiddenPlaylists[0]!));

        await waitFor(() => expect(result.current.feedback.message).toBe('Nao encontrado'));
        expect(result.current.feedback.severity).toBe('error');
    });

    it('falls back to the translated message when the failure has no backend text', async () => {
        const { result } = renderProviderAt('/videos/series');
        await waitFor(() => expect(result.current.seriesPlaylists).toHaveLength(1));
        mockedApi.put.mockRejectedValue(new Error('network'));

        act(() => result.current.hidePlaylist(result.current.seriesPlaylists[0]!));

        await waitFor(() =>
            expect(result.current.feedback.message).toBe('VIDEO_PLAYLIST_HIDE_ERROR')
        );
    });

    it('pages the hidden playlists through loadMoreHiddenPlaylists', async () => {
        mockedApi.get.mockImplementation((url: string, config?: { params?: { page: number } }) => {
            if (url === '/video/playlists/hidden') {
                return Promise.resolve(
                    config?.params?.page === 2
                        ? pageOf([buildPlaylist(9, 'segunda', true)])
                        : pageOf([buildPlaylist(8, 'oculta', true)], true)
                );
            }
            return Promise.resolve(pageOf([]));
        });
        const { result } = renderProviderAt('/videos/hidden');
        await waitFor(() => expect(result.current.hasMoreHiddenPlaylists).toBe(true));

        act(() => result.current.loadMoreHiddenPlaylists());

        await waitFor(() =>
            expect(result.current.hiddenPlaylists.map((p) => p.id)).toEqual([8, 9])
        );
    });
});
