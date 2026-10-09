import { fireEvent, screen, waitFor } from '@testing-library/react';
import { render } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import I18nProvider from '@/components/i18n/provider';
import * as musicService from '@/service/music';
import AlbumsView from './AlbumsView';
import ArtistsView from './ArtistsView';
import FoldersView from './FoldersView';
import GenresView from './GenresView';

jest.mock('@/service/music');
jest.mock('@/service/playlist');
jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => ({
        replaceQueue: jest.fn(),
        currentTrack: null,
        isPlaying: false,
    }),
}));

const mockedMusic = jest.mocked(musicService);

const pageOf = <ItemType,>(items: ItemType[], page: number, hasNext: boolean) => ({
    items,
    pagination: { page, page_size: 50, has_next: hasNext, has_prev: page > 1 },
});

const trackNamed = (id: number, metadata: Record<string, string>) => ({
    id,
    name: `track-${id}.mp3`,
    path: `/music/track-${id}.mp3`,
    metadata: { title: `Track ${id}`, ...metadata },
});

const renderViewAt = (view: React.ReactNode, search = '/') =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <I18nProvider>
                <MemoryRouter initialEntries={[search]}>{view}</MemoryRouter>
            </I18nProvider>
        </QueryClientProvider>
    );

const renderView = (view: React.ReactNode) => renderViewAt(view);

type CollectionCase = {
    name: string;
    View: React.ComponentType;
    listMock: jest.Mock;
    tracksMock: jest.Mock;
    firstGroup: unknown;
    secondGroup: unknown;
    firstGroupLabel: string;
    secondGroupLabel: string;
    emptyKey: string;
    deepLinkKey: string;
    deepLinkTrackMetadata: Record<string, string>;
    deepLinkTitle: string;
};

const collectionCases: CollectionCase[] = [
    {
        name: 'ArtistsView',
        View: ArtistsView,
        listMock: mockedMusic.getMusicArtists as jest.Mock,
        tracksMock: mockedMusic.getMusicByArtist as jest.Mock,
        firstGroup: { key: 'a', artist: 'Artist A', track_count: 7, album_count: 1 },
        secondGroup: { key: 'b', artist: 'Artist B', track_count: 2, album_count: 1 },
        firstGroupLabel: 'Artist A',
        secondGroupLabel: 'Artist B',
        emptyKey: 'MUSIC_ARTISTS_EMPTY',
        deepLinkKey: 'unloaded-artist',
        deepLinkTrackMetadata: { artist: 'Unloaded Artist' },
        deepLinkTitle: 'Unloaded Artist',
    },
    {
        name: 'AlbumsView',
        View: AlbumsView,
        listMock: mockedMusic.getMusicAlbums as jest.Mock,
        tracksMock: mockedMusic.getMusicByAlbum as jest.Mock,
        firstGroup: { key: 'a', album: 'Album A', artist: 'X', year: '2000', track_count: 7 },
        secondGroup: { key: 'b', album: 'Album B', artist: 'Y', year: '2001', track_count: 2 },
        firstGroupLabel: 'Album A',
        secondGroupLabel: 'Album B',
        emptyKey: 'MUSIC_ALBUMS_EMPTY',
        deepLinkKey: 'unloaded-album',
        deepLinkTrackMetadata: { album: 'Unloaded Album', artist: 'Someone' },
        deepLinkTitle: 'Unloaded Album',
    },
    {
        name: 'GenresView',
        View: GenresView,
        listMock: mockedMusic.getMusicGenres as jest.Mock,
        tracksMock: mockedMusic.getMusicByGenre as jest.Mock,
        firstGroup: { key: 'a', genre: 'Genre A', track_count: 7 },
        secondGroup: { key: 'b', genre: 'Genre B', track_count: 2 },
        firstGroupLabel: 'Genre A',
        secondGroupLabel: 'Genre B',
        emptyKey: 'MUSIC_GENRES_EMPTY',
        deepLinkKey: 'unloaded-genre',
        deepLinkTrackMetadata: { genre: 'Unloaded Genre' },
        deepLinkTitle: 'Unloaded Genre',
    },
    {
        name: 'FoldersView',
        View: FoldersView,
        listMock: mockedMusic.getMusicFolders as jest.Mock,
        tracksMock: mockedMusic.getMusicByFolder as jest.Mock,
        firstGroup: { folder: '/music/folder-a', track_count: 7 },
        secondGroup: { folder: '/music/folder-b', track_count: 2 },
        firstGroupLabel: 'folder-a',
        secondGroupLabel: 'folder-b',
        emptyKey: 'MUSIC_FOLDERS_EMPTY',
        deepLinkKey: '/music/unloaded-folder',
        deepLinkTrackMetadata: {},
        deepLinkTitle: 'unloaded-folder',
    },
];

const queryParamByView: Record<string, string> = {
    ArtistsView: 'artist',
    AlbumsView: 'album',
    GenresView: 'genre',
    FoldersView: 'folder',
};

describe.each(collectionCases)('$name', (collectionCase) => {
    const { View, listMock, tracksMock } = collectionCase;

    beforeEach(() => {
        jest.clearAllMocks();
        listMock.mockResolvedValue(pageOf([], 1, false));
        tracksMock.mockResolvedValue(pageOf([], 1, false));
    });

    it('renders the empty state when there are no groups', async () => {
        renderView(<View />);

        expect(await screen.findByText(collectionCase.emptyKey)).toBeInTheDocument();
    });

    it('shows the backend message verbatim and retries', async () => {
        listMock.mockRejectedValueOnce({ response: { data: { error: 'falha do servidor' } } });
        listMock.mockResolvedValueOnce(pageOf([collectionCase.firstGroup], 1, false));
        renderView(<View />);

        expect(await screen.findByText('falha do servidor')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: /Tentar novamente|TRY_AGAIN/ }));

        expect(await screen.findByText(collectionCase.firstGroupLabel)).toBeInTheDocument();
        expect(listMock).toHaveBeenCalledTimes(2);
    });

    it('fetches the next page through the load more sentinel', async () => {
        listMock.mockResolvedValueOnce(pageOf([collectionCase.firstGroup], 1, true));
        listMock.mockResolvedValueOnce(pageOf([collectionCase.secondGroup], 2, false));
        renderView(<View />);

        expect(await screen.findByText(collectionCase.firstGroupLabel)).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: /LOAD_MORE|Load more/i }));

        expect(await screen.findByText(collectionCase.secondGroupLabel)).toBeInTheDocument();
        expect(listMock).toHaveBeenCalledTimes(2);
    });

    it('opens a deep link whose group is not in the loaded pages', async () => {
        listMock.mockResolvedValue(pageOf([collectionCase.firstGroup], 1, true));
        tracksMock.mockResolvedValue(
            pageOf([trackNamed(1, collectionCase.deepLinkTrackMetadata)], 1, false)
        );
        renderViewAt(
            <View />,
            `/?${queryParamByView[collectionCase.name]}=${encodeURIComponent(collectionCase.deepLinkKey)}`
        );

        expect(await screen.findByText(collectionCase.deepLinkTitle)).toBeInTheDocument();
        expect(await screen.findByText('Track 1')).toBeInTheDocument();
        expect(await screen.findByText('1 MUSIC_TRACKS_COUNT')).toBeInTheDocument();
    });

    it('uses the group total in the detail header and fetches more tracks with the sentinel', async () => {
        listMock.mockResolvedValue(pageOf([collectionCase.firstGroup], 1, false));
        tracksMock.mockResolvedValueOnce(pageOf([trackNamed(1, {})], 1, true));
        tracksMock.mockResolvedValueOnce(pageOf([trackNamed(2, {})], 2, false));
        const firstGroupKey =
            (collectionCase.firstGroup as { key?: string; folder?: string }).key ??
            (collectionCase.firstGroup as { folder: string }).folder;
        renderViewAt(
            <View />,
            `/?${queryParamByView[collectionCase.name]}=${encodeURIComponent(firstGroupKey)}`
        );

        expect(await screen.findByText('Track 1')).toBeInTheDocument();
        await waitFor(() => expect(screen.getByText('7 MUSIC_TRACKS_COUNT')).toBeInTheDocument());
        fireEvent.click(screen.getByRole('button', { name: /LOAD_MORE|Load more/i }));

        expect(await screen.findByText('Track 2')).toBeInTheDocument();
    });

    it('shows the tracks error with retry in the detail', async () => {
        tracksMock.mockRejectedValueOnce({ response: { data: { error: 'faixas indisponiveis' } } });
        tracksMock.mockResolvedValueOnce(pageOf([trackNamed(1, {})], 1, false));
        renderViewAt(
            <View />,
            `/?${queryParamByView[collectionCase.name]}=${encodeURIComponent('/music/x')}`
        );

        expect(await screen.findByText('faixas indisponiveis')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: /Tentar novamente|TRY_AGAIN/ }));

        expect(await screen.findByText('Track 1')).toBeInTheDocument();
    });
});
