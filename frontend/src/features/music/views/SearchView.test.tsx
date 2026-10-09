import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import SearchView from './SearchView';

const mockSearchMusicTracks = jest.fn();
const mockReplaceQueue = jest.fn();

jest.mock('@/service/music', () => ({
    searchMusicTracks: (...args: unknown[]) => mockSearchMusicTracks(...args),
}));

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => ({ currentTrack: null, isPlaying: false }),
    useOptionalGlobalMusic: () => ({ replaceQueue: mockReplaceQueue }),
}));

jest.mock('@/features/music/components/AddToPlaylistMenu', () => () => null);

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) =>
            params ? `${key}:${JSON.stringify(params)}` : key,
    }),
}));

const buildTrack = (id: number, title: string) => ({
    id,
    name: title,
    path: `/m/${title}.mp3`,
    format: '.mp3',
    size: 1,
    metadata: { title, artist: 'Queen', length: 120 },
});

const renderSearchView = (initialEntry: string) => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={queryClient}>
            <MemoryRouter initialEntries={[initialEntry]}>
                <SearchView />
            </MemoryRouter>
        </QueryClientProvider>
    );
};

describe('SearchView', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('asks for a search term and does not hit the backend when q is missing', () => {
        renderSearchView('/music/search');

        expect(screen.getByText('MUSIC_SEARCH_PROMPT')).toBeInTheDocument();
        expect(mockSearchMusicTracks).not.toHaveBeenCalled();
    });

    it('lists the tracks found for q and plays the clicked one with the search context', async () => {
        mockSearchMusicTracks.mockResolvedValue({
            items: [buildTrack(1, 'Bohemian'), buildTrack(2, 'Another')],
            pagination: { page: 1, page_size: 50, has_next: false, has_prev: false },
        });
        renderSearchView('/music/search?q=queen');

        const secondTrack = await screen.findByLabelText('play Another');
        expect(mockSearchMusicTracks).toHaveBeenCalledWith('queen', 1, 50);

        fireEvent.click(secondTrack);

        expect(mockReplaceQueue).toHaveBeenCalledWith(
            expect.arrayContaining([expect.objectContaining({ id: 2 })]),
            1,
            expect.objectContaining({
                href: '/music/search?q=queen',
                labelKey: 'MUSIC_PLAYBACK_CONTEXT_SEARCH',
            })
        );
    });

    it('shows the empty state when nothing matches', async () => {
        mockSearchMusicTracks.mockResolvedValue({
            items: [],
            pagination: { page: 1, page_size: 50, has_next: false, has_prev: false },
        });
        renderSearchView('/music/search?q=zzz');

        await waitFor(() => expect(screen.getByText('MUSIC_SEARCH_EMPTY')).toBeInTheDocument());
    });
});
