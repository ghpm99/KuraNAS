import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { apiBase } from '@/service';
import ArtistPage from './ArtistPage';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn() },
}));

const mockedReplaceQueue = jest.fn();

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => ({
        replaceQueue: mockedReplaceQueue,
        currentTrack: null,
        isPlaying: false,
    }),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock };

const artistKey = 'the band';
const artistPath = `/music/library/artists/${encodeURIComponent(artistKey)}`;

const artistSummary = {
    key: artistKey,
    name: 'The Band',
    track_count: 5,
    album_count: 2,
    total_length_seconds: 1800,
};

const albumGroup = (key: string, name: string, year: string) => ({
    key,
    album: name,
    artist: 'The Band',
    year,
    track_count: 2,
});

const pageOf = (items: unknown[], page: number, hasNext: boolean) => ({
    items,
    pagination: { page, page_size: 50, has_next: hasNext, has_prev: page > 1 },
});

const trackNamed = (id: number) => ({
    id,
    name: `track-${id}.mp3`,
    path: `/music/track-${id}.mp3`,
    metadata: { title: `Track ${id}`, artist: 'The Band', length: 100 },
});

const LocationProbe = () => {
    const location = useLocation();
    return <span data-testid="location">{`${location.pathname}${location.search}`}</span>;
};

const renderPage = () =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <MemoryRouter initialEntries={['/music/artists']}>
                <Routes>
                    <Route
                        path="/music/artists"
                        element={<ArtistPage artistKey={artistKey} onBack={jest.fn()} />}
                    />
                    <Route path="/music/albums" element={<LocationProbe />} />
                </Routes>
            </MemoryRouter>
        </QueryClientProvider>
    );

const respondWith = (albumPages: unknown[]) => {
    const albumPagesQueue = [...albumPages];
    mockedApi.get.mockImplementation((url: string) => {
        if (url === artistPath) return Promise.resolve({ data: artistSummary });
        if (url === `${artistPath}/albums`) {
            return Promise.resolve({ data: albumPagesQueue.shift() ?? pageOf([], 1, false) });
        }
        if (url === `${artistPath}/tracks`) {
            return Promise.resolve({ data: pageOf([trackNamed(1), trackNamed(2)], 1, false) });
        }
        if (url === `${artistPath}/queue`) {
            return Promise.resolve({
                data: {
                    items: [1, 2].map((fileId) => ({
                        file_id: fileId,
                        name: `${fileId}.mp3`,
                        path: `/music/${fileId}.mp3`,
                        format: '.mp3',
                        title: `Queue ${fileId}`,
                        artist: 'The Band',
                        album: 'A',
                        length: 100,
                    })),
                    truncated: false,
                },
            });
        }
        return Promise.reject(new Error(`unexpected ${url}`));
    });
};

describe('ArtistPage (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('shows counts and total duration from the summary and lists all tracks', async () => {
        respondWith([pageOf([albumGroup('the band::first', 'First', '1999')], 1, false)]);
        renderPage();

        expect(await screen.findByText('The Band')).toBeInTheDocument();
        expect(await screen.findByText('5 MUSIC_TRACKS_COUNT')).toBeInTheDocument();
        expect(screen.getByText('2 MUSIC_ALBUMS')).toBeInTheDocument();
        expect(screen.getByText(/MUSIC_DURATION_MINUTES/)).toBeInTheDocument();
        expect(screen.getByText('MUSIC_ARTIST_ALL_TRACKS')).toBeInTheDocument();
        expect(await screen.findByText('Track 1')).toBeInTheDocument();
    });

    it('renders the artist albums grid and paginates it through the sentinel', async () => {
        respondWith([
            pageOf([albumGroup('the band::first', 'First', '1999')], 1, true),
            pageOf([albumGroup('the band::second', 'Second', '2004')], 2, false),
        ]);
        renderPage();

        expect(await screen.findByText('First')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: /LOAD_MORE|Load more/i }));

        expect(await screen.findByText('Second')).toBeInTheDocument();
        expect(mockedApi.get).toHaveBeenCalledWith(`${artistPath}/albums`, {
            params: { page: 2, page_size: 50 },
        });
    });

    it('opens an album with the album query param', async () => {
        respondWith([pageOf([albumGroup('the band::first', 'First', '1999')], 1, false)]);
        renderPage();

        fireEvent.click(await screen.findByText('First'));

        expect(await screen.findByTestId('location')).toHaveTextContent(
            '/music/albums?album=the%20band%3A%3Afirst'
        );
    });

    it('hides the albums section when the artist has no album groups', async () => {
        respondWith([]);
        renderPage();

        expect(await screen.findByText('Track 1')).toBeInTheDocument();
        expect(screen.queryByText('MUSIC_ALBUMS')).not.toBeInTheDocument();
    });

    it('plays and shuffles through the artist queue endpoint', async () => {
        respondWith([]);
        renderPage();
        await screen.findByText('Track 1');

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_PLAY_ALL' }));
        await waitFor(() => expect(mockedReplaceQueue).toHaveBeenCalledTimes(1));
        expect(mockedApi.get).toHaveBeenCalledWith(`${artistPath}/queue`);

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_SHUFFLE_ALL' }));
        await waitFor(() => expect(mockedReplaceQueue).toHaveBeenCalledTimes(2));
    });
});
