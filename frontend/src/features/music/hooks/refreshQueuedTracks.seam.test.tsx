import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import { apiBase } from '@/service';
import type { IMusicData } from '@/types/music';
import useTrackMetadataForm from '@/features/music/components/trackMetadata/useTrackMetadataForm';
import useBatchTrackMetadataForm from '@/features/music/components/trackMetadata/useBatchTrackMetadataForm';
import useRenameGenre from '@/features/music/components/genreRename/useRenameGenre';
import {
    GlobalMusicProvider,
    useGlobalMusic,
} from '@/features/music/providers/GlobalMusicProvider';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), put: jest.fn() },
}));

const mockEnqueueSnackbar = jest.fn();
jest.mock('notistack', () => ({
    useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

jest.mock('@/features/music/providers/globalMusic/useAudioEngine', () => ({
    __esModule: true,
    default: () => ({
        audioRef: { current: null },
        loadAndPlayUrl: jest.fn(),
        loadUrlPaused: jest.fn(),
        preloadUrl: jest.fn(),
        canPlayType: jest.fn().mockReturnValue('maybe'),
        getPositionSeconds: jest.fn().mockReturnValue(0),
        togglePlayPause: jest.fn(),
        seek: jest.fn(),
        setVolume: jest.fn(),
        stop: jest.fn(),
        isPlaying: true,
        currentTime: 0,
        duration: 0,
        volume: 1,
    }),
}));

jest.mock('@/features/music/providers/globalMusic/useMusicStateSync', () => ({
    __esModule: true,
    default: () => ({ syncState: jest.fn() }),
}));

jest.mock('@/features/music/providers/globalMusic/usePlayReporting', () => ({
    __esModule: true,
    default: jest.fn(),
}));

jest.mock('@/features/music/providers/globalMusic/useMusicQueueHydration', () => ({
    __esModule: true,
    default: () => ({ hasSettled: true }),
}));

jest.mock('@/features/music/providers/globalMusic/useMusicQueuePersistence', () => ({
    __esModule: true,
    default: jest.fn(),
}));

jest.mock('@/components/providers/settingsProvider/settingsContext', () => ({
    __esModule: true,
    useSettings: () => ({
        settings: { players: { remember_music_queue: false } },
        isLoading: false,
    }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; put: jest.Mock };

const createTrack = (id: number, title: string): IMusicData =>
    ({ id, name: `track-${id}`, metadata: { title } }) as IMusicData;

const wrapper = ({ children }: { children?: ReactNode }) => (
    <QueryClientProvider
        client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
        <GlobalMusicProvider>{children}</GlobalMusicProvider>
    </QueryClientProvider>
);

const playingTitle = (player: ReturnType<typeof useGlobalMusic>) =>
    player.currentTrack?.metadata?.title;

const renderWithPlayingQueue = <Result,>(useScreenHook: () => Result) => {
    const rendered = renderHook(() => ({ player: useGlobalMusic(), screen: useScreenHook() }), {
        wrapper,
    });
    act(() => {
        rendered.result.current.player.replaceQueue(
            [createTrack(1, 'Other'), createTrack(2, 'Old title')],
            1
        );
    });
    return rendered;
};

describe('player queue refresh after metadata edits (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockResolvedValue({ data: { items: [] } });
    });

    it('shows the title returned by the PUT in the player after editing one track', async () => {
        mockedApi.put.mockResolvedValue({ data: createTrack(2, 'New title') });
        const { result } = renderWithPlayingQueue(() =>
            useTrackMetadataForm({ track: createTrack(2, 'Old title'), onSaved: jest.fn() })
        );

        act(() => result.current.screen.setTextualValue('title', 'New title'));
        act(() => result.current.screen.submit());

        await waitFor(() => expect(playingTitle(result.current.player)).toBe('New title'));
        expect(result.current.player.currentIndex).toBe(1);
        expect(result.current.player.isPlaying).toBe(true);
    });

    describe('batch edit', () => {
        const renderBatch = () =>
            renderWithPlayingQueue(() =>
                useBatchTrackMetadataForm({
                    selector: { fileIds: [1, 2] },
                    onClose: jest.fn(),
                })
            );

        const confirmBatch = (rendered: ReturnType<typeof renderBatch>) => {
            act(() => rendered.result.current.screen.setTextualValue('album', 'Album'));
            mockedApi.put.mockResolvedValue({ data: { matched: 2, updated: 2, failed: [] } });
            act(() => rendered.result.current.screen.confirm());
        };

        it('fetches the current track by id and shows its new title in the player', async () => {
            mockedApi.get.mockResolvedValue({ data: createTrack(2, 'Batch title') });
            const rendered = renderBatch();

            confirmBatch(rendered);

            await waitFor(() =>
                expect(playingTitle(rendered.result.current.player)).toBe('Batch title')
            );
            expect(mockedApi.get).toHaveBeenCalledWith('/music/library/tracks/2');
        });

        it('keeps the old title without a toast error when the fetch fails', async () => {
            mockedApi.get.mockRejectedValue(new Error('offline'));
            const rendered = renderBatch();

            confirmBatch(rendered);

            await waitFor(() =>
                expect(mockedApi.get).toHaveBeenCalledWith('/music/library/tracks/2')
            );
            expect(playingTitle(rendered.result.current.player)).toBe('Old title');
            expect(mockEnqueueSnackbar).not.toHaveBeenCalledWith(
                expect.anything(),
                expect.objectContaining({ variant: 'error' })
            );
        });
    });

    describe('genre rename', () => {
        const renderRename = () =>
            renderWithPlayingQueue(() =>
                useRenameGenre({
                    genreKey: 'rock',
                    genreName: 'Rock',
                    trackCount: 2,
                    onRenamed: jest.fn(),
                    onClose: jest.fn(),
                })
            );

        const confirmRename = (rendered: ReturnType<typeof renderRename>) => {
            mockedApi.put.mockResolvedValue({
                data: { key: 'metal', name: 'Metal', track_count: 2, merged: false },
            });
            act(() => rendered.result.current.screen.typeNewName('Metal'));
            act(() => rendered.result.current.screen.confirm());
        };

        it('fetches the current track by id and shows its new title in the player', async () => {
            mockedApi.get.mockResolvedValue({ data: createTrack(2, 'Renamed genre title') });
            const rendered = renderRename();

            confirmRename(rendered);

            await waitFor(() =>
                expect(playingTitle(rendered.result.current.player)).toBe('Renamed genre title')
            );
        });

        it('keeps the old title without a toast error when the fetch fails', async () => {
            mockedApi.get.mockRejectedValue(new Error('offline'));
            const rendered = renderRename();

            confirmRename(rendered);

            await waitFor(() =>
                expect(mockedApi.get).toHaveBeenCalledWith('/music/library/tracks/2')
            );
            expect(playingTitle(rendered.result.current.player)).toBe('Old title');
            expect(mockEnqueueSnackbar).not.toHaveBeenCalledWith(
                expect.anything(),
                expect.objectContaining({ variant: 'error' })
            );
        });
    });

    it('does nothing outside a GlobalMusicProvider', async () => {
        mockedApi.put.mockResolvedValue({ data: { matched: 1, updated: 1, failed: [] } });
        const queryWrapper = ({ children }: { children?: ReactNode }) => (
            <QueryClientProvider client={new QueryClient()}>{children}</QueryClientProvider>
        );
        const { result } = renderHook(
            () => useBatchTrackMetadataForm({ selector: { fileIds: [1] }, onClose: jest.fn() }),
            { wrapper: queryWrapper }
        );

        act(() => result.current.setTextualValue('album', 'Album'));
        act(() => result.current.confirm());

        await waitFor(() => expect(mockedApi.put).toHaveBeenCalled());
        expect(mockedApi.get).not.toHaveBeenCalledWith(expect.stringContaining('/library/tracks/'));
    });
});
