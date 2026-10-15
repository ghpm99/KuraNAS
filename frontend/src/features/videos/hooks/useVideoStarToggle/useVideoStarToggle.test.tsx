import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { apiBase } from '@/service';
import { useVideoStarToggle } from './useVideoStarToggle';

const enqueueSnackbar = jest.fn();

jest.mock('@/service', () => ({
    apiBase: { put: jest.fn() },
}));

jest.mock('notistack', () => ({
    useSnackbar: () => ({ enqueueSnackbar }),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApiPut = apiBase.put as jest.Mock;

const buildVideo = (id: number, starred = false) => ({
    id,
    name: `video-${id}.mkv`,
    path: `/videos/video-${id}.mkv`,
    format: 'mkv',
    starred,
});

const libraryKey = ['videos', 'library-movies', 'name'];
const sessionKey = ['video', 'playlist-items', 5];

const setup = (onStarredChange?: (videoId: number, isStarred: boolean) => void) => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    queryClient.setQueryData(libraryKey, { pages: [{ items: [buildVideo(7), buildVideo(8)] }] });
    queryClient.setQueryData(sessionKey, [{ id: 1, video: buildVideo(7) }]);
    const wrapper = ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
    const hook = renderHook(() => useVideoStarToggle({ onStarredChange }), { wrapper });
    const readLibraryVideo = (videoId: number) =>
        queryClient
            .getQueryData<{ pages: { items: ReturnType<typeof buildVideo>[] }[] }>(libraryKey)!
            .pages[0]!.items.find((video) => video.id === videoId);
    const readSessionVideo = () =>
        queryClient.getQueryData<{ video: ReturnType<typeof buildVideo> }[]>(sessionKey)![0]!.video;
    return { queryClient, hook, readLibraryVideo, readSessionVideo };
};

describe('useVideoStarToggle', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('mounts without cached video data', () => {
        const queryClient = new QueryClient();
        const wrapper = ({ children }: { children: ReactNode }) => (
            <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
        );

        const { result } = renderHook(() => useVideoStarToggle(), { wrapper });

        expect(result.current.isStarTogglePending).toBe(false);
    });

    it('stars optimistically in every cache holding the video and calls the starred endpoint', async () => {
        let resolveRequest: () => void = () => undefined;
        mockedApiPut.mockReturnValue(
            new Promise((resolve) => (resolveRequest = () => resolve({ data: {} })))
        );
        const onStarredChange = jest.fn();
        const { hook, readLibraryVideo, readSessionVideo } = setup(onStarredChange);

        act(() => hook.result.current.toggleStar(7, false));

        await waitFor(() => expect(readLibraryVideo(7)?.starred).toBe(true));
        expect(readSessionVideo().starred).toBe(true);
        expect(readLibraryVideo(8)?.starred).toBe(false);
        expect(onStarredChange).toHaveBeenCalledWith(7, true);
        expect(mockedApiPut).toHaveBeenCalledWith('/files/starred/7', { starred: true });

        await act(async () => resolveRequest());
    });

    it('unstars when the video was starred', async () => {
        mockedApiPut.mockResolvedValue({ data: {} });
        const { hook, readLibraryVideo, queryClient } = setup();
        queryClient.setQueryData(libraryKey, { pages: [{ items: [buildVideo(7, true)] }] });

        act(() => hook.result.current.toggleStar(7, true));

        await waitFor(() => expect(readLibraryVideo(7)?.starred).toBe(false));
        expect(mockedApiPut).toHaveBeenCalledWith('/files/starred/7', { starred: false });
    });

    it('invalidates the starred views once the toggle settles', async () => {
        mockedApiPut.mockResolvedValue({ data: {} });
        const { hook, queryClient } = setup();
        const invalidateQueries = jest.spyOn(queryClient, 'invalidateQueries');

        act(() => hook.result.current.toggleStar(7, false));

        await waitFor(() =>
            expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['home', 'favorites'] })
        );
    });

    it('rolls every cache back and shows the server message on failure', async () => {
        mockedApiPut.mockRejectedValue({ response: { data: { error: 'Servidor recusou' } } });
        const onStarredChange = jest.fn();
        const { hook, readLibraryVideo, readSessionVideo } = setup(onStarredChange);

        act(() => hook.result.current.toggleStar(7, false));

        await waitFor(() =>
            expect(enqueueSnackbar).toHaveBeenCalledWith('Servidor recusou', { variant: 'error' })
        );
        expect(readLibraryVideo(7)?.starred).toBe(false);
        expect(readSessionVideo().starred).toBe(false);
        expect(onStarredChange).toHaveBeenLastCalledWith(7, false);
    });

    it('falls back to the generic message when the server sends none', async () => {
        mockedApiPut.mockRejectedValue(new Error('boom'));
        const { hook } = setup();

        act(() => hook.result.current.toggleStar(7, false));

        await waitFor(() =>
            expect(enqueueSnackbar).toHaveBeenCalledWith('ERROR_FAVORITE_FAILED', {
                variant: 'error',
            })
        );
    });
});
