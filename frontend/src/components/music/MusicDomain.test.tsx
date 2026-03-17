import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import MusicDomainHeader from './MusicDomainHeader';
import MusicHomeScreen from './MusicHomeScreen';
import MusicSidebar from './MusicSidebar';

const mockUseGlobalMusic = jest.fn();
const mockUseQuery = useQuery as jest.Mock;
const mockGetPlaylistTracks = jest.fn();
const mockGetMusicByArtist = jest.fn();
const mockGetMusicByAlbum = jest.fn();

jest.mock('@/components/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => mockUseGlobalMusic(),
}));

jest.mock('@/utils/music', () => ({
    getMusicTitle: (m: any) => m.name ?? m.metadata?.title ?? '',
    getMusicArtist: (m: any) => m.metadata?.artist ?? 'Unknown Artist',
    musicMetadata: () => 'meta',
    formatMusicDuration: (s: number) =>
        `${Math.floor(s / 60)}:${String(Math.floor(s % 60)).padStart(2, '0')}`,
}));

jest.mock('@tanstack/react-query', () => ({
    useQuery: jest.fn(),
}));

jest.mock('@/service/playlist', () => ({
    getPlaylistTracks: (...args: any[]) => mockGetPlaylistTracks(...args),
}));

jest.mock('@/service/music', () => ({
    getMusicByArtist: (...args: any[]) => mockGetMusicByArtist(...args),
    getMusicByAlbum: (...args: any[]) => mockGetMusicByAlbum(...args),
    getMusicHomeCatalog: jest.fn(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) => {
            if (params?.context) {
                return `${key}:${params.context}`;
            }
            if (params?.count) {
                return `${key}:${params.count}`;
            }
            if (params?.name) {
                return `${key}:${params.name}`;
            }
            return key;
        },
    }),
}));

describe('components/music domain shell', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockUseQuery.mockReturnValue({
            data: {
                summary: {
                    total_tracks: 3,
                    total_artists: 2,
                    total_albums: 2,
                    total_genres: 1,
                    total_folders: 1,
                },
                playlists: [
                    {
                        id: 5,
                        name: 'Playlist A',
                        description: 'Desc',
                        track_count: 4,
                        is_system: false,
                        is_auto: false,
                        kind: 'manual',
                        source_key: '',
                    },
                ],
                artists: [
                    {
                        key: 'artist-a',
                        artist: 'Artist A',
                        album_count: 1,
                        track_count: 2,
                    },
                ],
                albums: [
                    {
                        key: 'artist-a::album-a',
                        album: 'Album A',
                        artist: 'Artist A',
                        year: '2024',
                        track_count: 2,
                    },
                ],
            },
            isLoading: false,
            status: 'success',
        });
        mockGetPlaylistTracks.mockResolvedValue({
            items: [
                {
                    file: {
                        id: 20,
                        name: 'Playlist Song',
                        metadata: { title: 'Playlist Song', artist: 'Playlist Artist' },
                    },
                },
            ],
        });
        mockGetMusicByArtist.mockResolvedValue({
            items: [
                {
                    id: 30,
                    name: 'Artist Song',
                    metadata: {
                        title: 'Artist Song',
                        artist: 'Artist A',
                        album: 'Album A',
                    },
                },
            ],
        });
        mockGetMusicByAlbum.mockResolvedValue({
            items: [
                {
                    id: 40,
                    name: 'Album Song',
                    metadata: {
                        title: 'Album Song',
                        artist: 'Artist A',
                        album: 'Album A',
                    },
                },
            ],
        });
        mockUseGlobalMusic.mockReturnValue({
            currentIndex: 0,
            currentTrack: { id: 10, name: 'Song A' },
            hasQueue: true,
            playbackContext: {
                labelKey: 'MUSIC_PLAYBACK_CONTEXT_ALBUM',
                labelParams: { name: 'Album A' },
                href: '/music/albums',
            },
            queue: [{ id: 10 }, { id: 11, name: 'Song B' }],
            replaceQueue: jest.fn(),
            toggleQueue: jest.fn(),
        });
    });

    it('renders contextual header and active sidebar item from route', () => {
        render(
            <MemoryRouter initialEntries={['/music/albums']}>
                <MusicDomainHeader />
                <MusicSidebar />
            </MemoryRouter>
        );

        expect(screen.getByRole('heading', { name: 'MUSIC_ALBUMS' })).toBeInTheDocument();
        expect(screen.getAllByText('MUSIC_ALBUMS_DESCRIPTION')[0]).toBeInTheDocument();
        expect(screen.getByRole('link', { name: /MUSIC_ALBUMS/i })).toHaveAttribute(
            'href',
            '/music/albums'
        );
    });

    it('renders music home with queue, playback context, and content sections', () => {
        const toggleQueue = jest.fn();
        mockUseGlobalMusic.mockReturnValue({
            currentIndex: 0,
            currentTrack: { id: 10, name: 'Song A' },
            hasQueue: true,
            playbackContext: {
                labelKey: 'MUSIC_PLAYBACK_CONTEXT_ALBUM',
                labelParams: { name: 'Album A' },
                href: '/music/albums',
            },
            queue: [{ id: 10 }, { id: 11, name: 'Song B' }],
            replaceQueue: jest.fn(),
            toggleQueue,
        });

        render(
            <MemoryRouter initialEntries={['/music']}>
                <MusicHomeScreen />
            </MemoryRouter>
        );

        expect(screen.getByText('MUSIC_HOME_QUEUE_READY')).toBeInTheDocument();
        expect(screen.getAllByText('Song A').length).toBeGreaterThan(0);
        expect(screen.getByText('MUSIC_HOME_QUEUE_COUNT:2')).toBeInTheDocument();
        expect(screen.getByText('3')).toBeInTheDocument();
        expect(
            screen.getByText('MUSIC_PLAYBACK_FROM:MUSIC_PLAYBACK_CONTEXT_ALBUM:Album A')
        ).toBeInTheDocument();
        expect(screen.getByText('MUSIC_HOME_FEATURED_PLAYLISTS')).toBeInTheDocument();
        expect(screen.getByText('Playlist A')).toBeInTheDocument();
        expect(screen.getAllByText('Artist A').length).toBeGreaterThan(0);
        expect(screen.getAllByText('Album A').length).toBeGreaterThan(0);
        expect(screen.getByRole('link', { name: 'MUSIC_HOME_RETURN_TO_CONTEXT' })).toHaveAttribute(
            'href',
            '/music/albums'
        );
        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_HOME_OPEN_QUEUE' }));
        expect(toggleQueue).toHaveBeenCalled();
    });

    it('starts contextual playback from playlist, artist, and album cards', async () => {
        const replaceQueue = jest.fn();
        mockUseGlobalMusic.mockReturnValue({
            currentIndex: 0,
            currentTrack: { id: 10, name: 'Song A' },
            hasQueue: true,
            playbackContext: undefined,
            queue: [{ id: 10 }, { id: 11, name: 'Song B' }],
            replaceQueue,
            toggleQueue: jest.fn(),
        });

        render(
            <MemoryRouter initialEntries={['/music']}>
                <MusicHomeScreen />
            </MemoryRouter>
        );

        const playButtons = screen.getAllByRole('button', {
            name: 'MUSIC_HOME_PLAY_NOW',
        });
        fireEvent.click(playButtons[0]!);
        fireEvent.click(playButtons[1]!);
        fireEvent.click(playButtons[2]!);

        await waitFor(() => {
            expect(mockGetPlaylistTracks).toHaveBeenCalledWith(5, 1, 200);
            expect(mockGetMusicByArtist).toHaveBeenCalledWith('artist-a', 1, 200);
            expect(mockGetMusicByAlbum).toHaveBeenCalledWith('artist-a::album-a', 1, 200);
            expect(replaceQueue).toHaveBeenCalledWith(
                [expect.objectContaining({ id: 20 })],
                0,
                expect.objectContaining({ kind: 'playlist' })
            );
            expect(replaceQueue).toHaveBeenCalledWith(
                [expect.objectContaining({ id: 30 })],
                0,
                expect.objectContaining({ kind: 'artist' })
            );
            expect(replaceQueue).toHaveBeenCalledWith(
                [expect.objectContaining({ id: 40 })],
                0,
                expect.objectContaining({ kind: 'album' })
            );
        });
    });

    it('renders music home empty state while catalog is loading', () => {
        mockUseGlobalMusic.mockReturnValue({
            currentIndex: undefined,
            currentTrack: undefined,
            hasQueue: false,
            playbackContext: undefined,
            queue: [],
            replaceQueue: jest.fn(),
            toggleQueue: jest.fn(),
        });
        mockUseQuery.mockReturnValue({
            data: {
                summary: {
                    total_tracks: 0,
                    total_artists: 0,
                    total_albums: 0,
                    total_genres: 0,
                    total_folders: 0,
                },
                playlists: [],
                artists: [],
                albums: [],
            },
            isLoading: true,
            status: 'pending',
        });

        render(
            <MemoryRouter initialEntries={['/music']}>
                <MusicHomeScreen />
            </MemoryRouter>
        );

        expect(screen.getByText('MUSIC_HOME_QUEUE_EMPTY')).toBeInTheDocument();
        expect(screen.getByText('MUSIC_HOME_QUEUE_EMPTY_STATE')).toBeInTheDocument();
        expect(screen.getByText('MUSIC_HOME_LIBRARY_LOADING')).toBeInTheDocument();
        expect(screen.getByText('LOADING')).toBeInTheDocument();
    });
});
