import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter, useLocation } from 'react-router-dom';
import GlobalSearchProvider from './GlobalSearchProvider';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: {
        get: jest.fn(),
    },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) =>
            params ? `${key}:${JSON.stringify(params)}` : key,
    }),
}));

const mockReplaceQueue = jest.fn();
let mockHasGlobalMusic = true;

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useOptionalGlobalMusic: () =>
        mockHasGlobalMusic ? { replaceQueue: mockReplaceQueue } : undefined,
}));

const mockedApi = apiBase as unknown as { get: jest.Mock };

const trackResult = {
    file_id: 7,
    title: 'Time',
    artist: 'Pink Floyd',
    album: 'The Dark Side',
    album_key: 'pink floyd::the dark side',
    duration: 413,
    path: '/Music/time.mp3',
};

const searchResponse = {
    query: 'time',
    files: [],
    folders: [],
    artists: [],
    albums: [],
    playlists: [],
    videos: [],
    images: [],
    tracks: [trackResult],
};

const LocationProbe = () => {
    const location = useLocation();
    return <span data-testid="location">{`${location.pathname}${location.search}`}</span>;
};

const renderProvider = () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <MemoryRouter>
                <GlobalSearchProvider>
                    <LocationProbe />
                </GlobalSearchProvider>
            </MemoryRouter>
        </QueryClientProvider>
    );
};

const openAndType = (text: string) => {
    fireEvent.keyDown(window, { key: 'k', ctrlKey: true });
    fireEvent.change(screen.getByRole('combobox'), { target: { value: text } });
};

describe('components/search/GlobalSearchProvider tracks', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockHasGlobalMusic = true;
        mockedApi.get.mockResolvedValue({ data: searchResponse });
    });

    it('lists tracks in their own group and plays the clicked track through the global player', async () => {
        renderProvider();
        openAndType('time');

        expect(await screen.findByText('GLOBAL_SEARCH_SECTION_TRACKS')).toBeInTheDocument();
        expect(screen.getByText('Pink Floyd · The Dark Side')).toBeInTheDocument();
        expect(screen.getByText('6:53')).toBeInTheDocument();

        fireEvent.click(screen.getByText('Time'));

        expect(mockReplaceQueue).toHaveBeenCalledTimes(1);
        const [queuedTracks, startIndex, playbackContext] = mockReplaceQueue.mock.calls[0];
        expect(queuedTracks).toHaveLength(1);
        expect(queuedTracks[0]).toEqual(
            expect.objectContaining({ id: 7, path: '/Music/time.mp3', format: '.mp3' })
        );
        expect(startIndex).toBe(0);
        expect(playbackContext).toEqual(expect.objectContaining({ kind: 'album' }));
        await waitFor(() => expect(screen.queryByRole('combobox')).not.toBeInTheDocument());
    });

    it('plays the best result when Enter is pressed', async () => {
        renderProvider();
        openAndType('time');
        await screen.findByText('GLOBAL_SEARCH_SECTION_TRACKS');

        fireEvent.keyDown(screen.getByRole('combobox'), { key: 'Enter' });

        expect(mockReplaceQueue).toHaveBeenCalledTimes(1);
    });

    it('opens the album from the secondary action without playing', async () => {
        renderProvider();
        openAndType('time');

        fireEvent.click(await screen.findByText('GLOBAL_SEARCH_OPEN_ALBUM'));

        expect(mockReplaceQueue).not.toHaveBeenCalled();
        expect(screen.getByTestId('location')).toHaveTextContent(
            '/music/albums?album=pink+floyd%3A%3Athe+dark+side'
        );
    });

    it('opens the album with Shift+Enter', async () => {
        renderProvider();
        openAndType('time');
        await screen.findByText('GLOBAL_SEARCH_SECTION_TRACKS');

        fireEvent.keyDown(screen.getByRole('combobox'), { key: 'Enter', shiftKey: true });

        expect(mockReplaceQueue).not.toHaveBeenCalled();
        expect(screen.getByTestId('location')).toHaveTextContent('/music/albums');
    });

    it('offers no album action for a track without album key', async () => {
        mockedApi.get.mockResolvedValue({
            data: { ...searchResponse, tracks: [{ ...trackResult, album: '', album_key: '' }] },
        });
        renderProvider();
        openAndType('time');
        await screen.findByText('GLOBAL_SEARCH_SECTION_TRACKS');

        expect(screen.queryByText('GLOBAL_SEARCH_OPEN_ALBUM')).not.toBeInTheDocument();
        fireEvent.keyDown(screen.getByRole('combobox'), { key: 'Enter', shiftKey: true });
        expect(screen.getByRole('combobox')).toBeInTheDocument();
    });

    it('opens the file location when no global player is available', async () => {
        mockHasGlobalMusic = false;
        renderProvider();
        openAndType('time');

        fireEvent.click(await screen.findByText('Time'));

        expect(mockReplaceQueue).not.toHaveBeenCalled();
        expect(screen.getByTestId('location')).toHaveTextContent('/files');
    });

    it('renders results from an older backend that sends no tracks', async () => {
        const { tracks: ignoredTracks, ...legacyResponse } = searchResponse;
        expect(ignoredTracks).toHaveLength(1);
        mockedApi.get.mockResolvedValue({
            data: {
                ...legacyResponse,
                files: [
                    {
                        id: 1,
                        name: 'notes.txt',
                        path: '/notes.txt',
                        parent_path: '/',
                        format: '.txt',
                        starred: false,
                    },
                ],
            },
        });
        renderProvider();
        openAndType('notes');

        expect(await screen.findByRole('option', { name: /notes\s*\.txt/ })).toBeInTheDocument();
        expect(screen.queryByText('GLOBAL_SEARCH_SECTION_TRACKS')).not.toBeInTheDocument();
    });
});
