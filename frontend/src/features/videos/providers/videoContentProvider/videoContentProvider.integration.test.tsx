import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import type { ReactNode } from 'react';
import { VideoContentProvider, useVideoContentProvider } from './videoContentProvider';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real VideoContentProvider +
// service/videoPlayback.ts run. Rendered at the playlist detail route so the
// playlist-scoped commands resolve a selected playlist, each command asserts the
// exact endpoint/payload the backend video handlers decode.
jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}));

jest.mock('react-router-dom', () => {
    const actual = jest.requireActual('react-router-dom');
    return { ...actual, useNavigate: () => jest.fn() };
});

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as {
    get: jest.Mock;
    post: jest.Mock;
    put: jest.Mock;
    delete: jest.Mock;
};

// name 'Série' slugifies to 'serie', matching the route below.
const playlist = {
    id: 2,
    type: 'series',
    source_path: '/videos/serie',
    name: 'Série',
    is_hidden: false,
    is_auto: true,
    group_mode: 'folder',
    classification: 'series',
    item_count: 2,
    cover_video_id: 10,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    last_played_at: null,
    items: [
        { id: 1, order_index: 0, video: { id: 10 } },
        { id: 2, order_index: 1, video: { id: 20 } },
    ],
};

const wrapper = ({ children }: { children: ReactNode }) => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return (
        <QueryClientProvider client={client}>
            <MemoryRouter initialEntries={['/videos/playlists/serie']}>
                <VideoContentProvider>{children}</VideoContentProvider>
            </MemoryRouter>
        </QueryClientProvider>
    );
};

const renderProvider = () => renderHook(() => useVideoContentProvider(), { wrapper });

describe('features/videos/videoContentProvider (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockImplementation((url: string) => {
            if (url.startsWith('/video/playlists/memberships'))
                return Promise.resolve({ data: [] });
            if (url === '/video/playlists/') return Promise.resolve({ data: [playlist] });
            if (/^\/video\/playlists\/\d+\/items$/.test(url)) {
                return Promise.resolve({
                    data: {
                        items: playlist.items,
                        pagination: { page: 1, page_size: 50, has_next: false, has_prev: false },
                    },
                });
            }
            if (url === '/video/continue') {
                return Promise.resolve({
                    data: [
                        {
                            video: { id: 10 },
                            position_seconds: 5,
                            duration_seconds: 50,
                            updated_at: '',
                        },
                        {
                            video: { id: 20 },
                            position_seconds: 6,
                            duration_seconds: 60,
                            updated_at: '',
                        },
                    ],
                });
            }
            if (url === '/video/catalog/home') return Promise.resolve({ data: { sections: [] } });
            if (url.startsWith('/video/library/files')) {
                return Promise.resolve({ data: { items: [], pagination: { has_next: false } } });
            }
            return Promise.resolve({
                data: { items: [], pagination: { has_next: false }, sections: [] },
            });
        });
        mockedApi.post.mockResolvedValue({ data: undefined });
        mockedApi.put.mockResolvedValue({ data: undefined });
        mockedApi.delete.mockResolvedValue({ data: undefined });
    });

    it('loads the selected playlist items from the paginated items endpoint', async () => {
        const { result } = renderProvider();
        await waitFor(() => expect(result.current.selectedPlaylistDetail?.items).toHaveLength(2));

        expect(mockedApi.get).toHaveBeenCalledWith('/video/playlists/2/items', {
            params: { page: 1, page_size: 50 },
        });
        expect(mockedApi.get).not.toHaveBeenCalledWith('/video/playlists/2');
    });

    it('addVideoFromLibrary POSTs the video id to the playlist videos endpoint', async () => {
        const { result } = renderProvider();
        await waitFor(() => expect(result.current.playlists.length).toBeGreaterThan(0));

        act(() => result.current.addVideoFromLibrary(100));

        await waitFor(() =>
            expect(mockedApi.post).toHaveBeenCalledWith('/video/playlists/2/videos', {
                video_id: 100,
            })
        );
    });

    it('renameSelectedPlaylist PUTs the new name to the playlist endpoint', async () => {
        const { result } = renderProvider();
        await waitFor(() => expect(result.current.selectedPlaylistSummary?.id).toBe(2));

        act(() => result.current.renameSelectedPlaylist('Temporada 2'));

        await waitFor(() =>
            expect(mockedApi.put).toHaveBeenCalledWith('/video/playlists/2', {
                name: 'Temporada 2',
            })
        );
    });

    it('removeVideoFromSelectedPlaylist DELETEs the video from the playlist', async () => {
        const { result } = renderProvider();
        await waitFor(() => expect(result.current.selectedPlaylistSummary?.id).toBe(2));

        act(() => result.current.removeVideoFromSelectedPlaylist(20));

        await waitFor(() =>
            expect(mockedApi.delete).toHaveBeenCalledWith('/video/playlists/2/videos/20')
        );
    });

    it('moveSelectedPlaylistItem PUTs the reordered items', async () => {
        const { result } = renderProvider();
        await waitFor(() => expect(result.current.selectedPlaylistDetail?.items?.length).toBe(2));

        act(() => result.current.moveSelectedPlaylistItem(0, 1));

        await waitFor(() =>
            expect(mockedApi.put).toHaveBeenCalledWith('/video/playlists/2/reorder', {
                items: [
                    { video_id: 10, order_index: 1 },
                    { video_id: 20, order_index: 0 },
                ],
            })
        );
    });

    it('setVideoWatched PUTs the watched flag to the progress endpoint', async () => {
        const { result } = renderProvider();
        await waitFor(() => expect(result.current.playlists.length).toBeGreaterThan(0));

        act(() => result.current.setVideoWatched(10, true));
        await waitFor(() =>
            expect(mockedApi.put).toHaveBeenCalledWith('/video/progress/10/watched', {
                watched: true,
            })
        );

        act(() => result.current.setVideoWatched(10, false));
        await waitFor(() =>
            expect(mockedApi.put).toHaveBeenCalledWith('/video/progress/10/watched', {
                watched: false,
            })
        );
    });

    it('setVideoWatched optimistically completes the playlist item and drops it from continue watching', async () => {
        const { result } = renderProvider();
        await waitFor(() => expect(result.current.selectedPlaylistDetail?.items?.length).toBe(2));
        await waitFor(() => expect(result.current.continueWatchingItems).toHaveLength(2));
        mockedApi.put.mockReturnValue(new Promise(() => undefined));

        act(() => result.current.setVideoWatched(10, true));

        await waitFor(() => {
            expect(result.current.continueWatchingItems.map((item) => item.video.id)).toEqual([20]);
            expect(result.current.selectedPlaylistDetail?.items[0]).toMatchObject({
                status: 'completed',
                progress_pct: 100,
            });
        });
    });

    it('setVideoWatched rolls back the optimistic update and reports the failure', async () => {
        const { result } = renderProvider();
        await waitFor(() => expect(result.current.continueWatchingItems).toHaveLength(2));
        mockedApi.put.mockRejectedValue(new Error('boom'));

        act(() => result.current.setVideoWatched(10, true));

        await waitFor(() => expect(result.current.feedback.message).toBe('VIDEO_WATCHED_ERROR'));
        await waitFor(() => expect(result.current.continueWatchingItems).toHaveLength(2));
    });
});
