import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import { apiBase } from '@/service';
import AlbumPage from './AlbumPage';

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

const albumKey = 'the band::double album';
const encodedAlbumKey = encodeURIComponent(albumKey);

const albumSummary = {
    key: albumKey,
    name: 'Double Album',
    artist: 'The Band',
    year: '2001',
    track_count: 3,
    total_length_seconds: 3725,
    disc_count: 2,
};

const trackOnDisc = (id: number, title: string, discNumber: string, trackNumber: string) => ({
    id,
    name: `${title}.mp3`,
    path: `/music/${title}.mp3`,
    metadata: {
        title,
        artist: 'The Band',
        album: 'Double Album',
        disc_number: discNumber,
        track_number: trackNumber,
        length: 200,
    },
});

const queueEntry = (fileId: number) => ({
    file_id: fileId,
    name: `${fileId}.mp3`,
    path: `/music/${fileId}.mp3`,
    format: '.mp3',
    title: `Queue ${fileId}`,
    artist: 'The Band',
    album: 'Double Album',
    length: 200,
});

const paginationOf = (items: unknown[]) => ({
    items,
    pagination: { page: 1, page_size: 50, has_next: false, has_prev: false },
});

const respondWith = (summary: unknown, tracks: unknown[]) => {
    mockedApi.get.mockImplementation((url: string) => {
        if (url === `/music/library/albums/${encodedAlbumKey}`) {
            return summary instanceof Error
                ? Promise.reject(summary)
                : Promise.resolve({ data: summary });
        }
        if (url === `/music/library/albums/${encodedAlbumKey}/tracks`) {
            return Promise.resolve({ data: paginationOf(tracks) });
        }
        if (url === `/music/library/albums/${encodedAlbumKey}/queue`) {
            return Promise.resolve({
                data: { items: [queueEntry(1), queueEntry(2), queueEntry(3)], truncated: false },
            });
        }
        return Promise.reject(new Error(`unexpected ${url}`));
    });
};

const renderPage = () =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <MemoryRouter>
                <AlbumPage albumKey={albumKey} onBack={jest.fn()} />
            </MemoryRouter>
        </QueryClientProvider>
    );

const discTracks = [
    trackOnDisc(1, 'Opening', '1/2', '1/2'),
    trackOnDisc(2, 'Closing', '1/2', '2/2'),
    trackOnDisc(3, 'Encore', '2/2', '7'),
];

describe('AlbumPage (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('shows the header with exact totals and an artist link from the summary endpoint', async () => {
        respondWith(albumSummary, discTracks);
        renderPage();

        expect(await screen.findByText('Double Album')).toBeInTheDocument();
        expect(await screen.findByText('3 MUSIC_TRACKS_COUNT')).toBeInTheDocument();
        expect(screen.getByText(/MUSIC_DURATION_HOURS_MINUTES/)).toBeInTheDocument();
        expect(screen.getByText(/· 2001/)).toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'The Band' })).toHaveAttribute(
            'href',
            '/music/artists?artist=the%20band'
        );
    });

    it('groups tracks by disc and shows the track number instead of the list position', async () => {
        respondWith(albumSummary, discTracks);
        renderPage();

        expect(await screen.findByText('Opening')).toBeInTheDocument();
        expect(
            screen.getAllByRole('heading', { level: 3 }).map((heading) => heading.textContent)
        ).toEqual(['MUSIC_DISC_HEADING', 'MUSIC_DISC_HEADING']);
        expect(screen.getByText('7')).toBeInTheDocument();
        expect(screen.queryByText('3')).not.toBeInTheDocument();
    });

    it('omits disc headings when the album has a single disc', async () => {
        respondWith({ ...albumSummary, disc_count: 1 }, [
            trackOnDisc(1, 'Opening', '', '1'),
            trackOnDisc(2, 'Closing', '', '2'),
        ]);
        renderPage();

        expect(await screen.findByText('Opening')).toBeInTheDocument();
        expect(screen.queryByRole('heading', { level: 3 })).not.toBeInTheDocument();
    });

    it('falls back to the track metadata when the summary is not found', async () => {
        respondWith(new Error('not found'), discTracks);
        renderPage();

        expect(await screen.findByText('Double Album')).toBeInTheDocument();
        expect(screen.queryByRole('link', { name: 'The Band' })).not.toBeInTheDocument();
        expect(screen.queryByText(/MUSIC_TRACKS_COUNT/)).not.toBeInTheDocument();
    });

    it('plays the album from the queue endpoint', async () => {
        respondWith(albumSummary, discTracks);
        renderPage();
        await screen.findByText('Opening');

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_PLAY_ALL' }));

        await waitFor(() => expect(mockedReplaceQueue).toHaveBeenCalledTimes(1));
        expect(mockedApi.get).toHaveBeenCalledWith(
            `/music/library/albums/${encodedAlbumKey}/queue`
        );
        const [queuedTracks, startIndex] = mockedReplaceQueue.mock.calls[0] as [
            { id: number }[],
            number,
        ];
        expect(queuedTracks.map((track) => track.id)).toEqual([1, 2, 3]);
        expect(startIndex).toBe(0);
    });

    it('shuffles the whole album from the same queue endpoint', async () => {
        respondWith(albumSummary, discTracks);
        renderPage();
        await screen.findByText('Opening');

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_SHUFFLE_ALL' }));

        await waitFor(() => expect(mockedReplaceQueue).toHaveBeenCalledTimes(1));
        const [queuedTracks] = mockedReplaceQueue.mock.calls[0] as [{ id: number }[]];
        expect(queuedTracks.map((track) => track.id).sort()).toEqual([1, 2, 3]);
    });

    it('starts the queue at the clicked track', async () => {
        respondWith(albumSummary, discTracks);
        renderPage();

        fireEvent.click(await screen.findByRole('button', { name: 'play Encore' }));

        await waitFor(() => expect(mockedReplaceQueue).toHaveBeenCalledTimes(1));
        expect(mockedReplaceQueue.mock.calls[0]?.[1]).toBe(2);
    });
});
