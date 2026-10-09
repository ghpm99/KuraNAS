import {
    getAlbumSearchRoute,
    getArtistSearchRoute,
    getFileSearchRoute,
    getFilesQuerySearchRoute,
    getImageSearchRoute,
    getImagesQuerySearchRoute,
    getMusicQuerySearchRoute,
    getVideosQuerySearchRoute,
    getPlaylistSearchRoute,
} from './searchResultRoutes';
import type { GlobalSearchPlaylistResult } from '@/service/search';

const buildVideoPlaylist = (
    overrides: Partial<GlobalSearchPlaylistResult>
): GlobalSearchPlaylistResult => ({
    scope: 'video',
    id: 7,
    name: 'Minha Série',
    description: 'custom',
    count: 3,
    classification: 'series',
    source_path: '',
    is_auto: false,
    ...overrides,
});

describe('searchResultRoutes', () => {
    it('builds the library search routes carrying the encoded term', () => {
        expect(getImagesQuerySearchRoute('praia & sol')).toEqual({
            pathname: '/images',
            search: '?q=praia+%26+sol',
        });
        expect(getVideosQuerySearchRoute('praia')).toEqual({
            pathname: '/videos/folders',
            search: '?q=praia',
        });
        expect(getMusicQuerySearchRoute('praia')).toEqual({
            pathname: '/music/search',
            search: '?q=praia',
        });
    });

    it('encodes file and folder names with reserved characters', () => {
        expect(getFileSearchRoute('/docs/a#b/c?d/100%.txt')).toBe(
            '/files/docs/a%23b/c%3Fd/100%25.txt'
        );
    });

    it('builds the files query route with an encoded term', () => {
        expect(getFilesQuerySearchRoute('a&b=c#d')).toEqual({
            pathname: '/files',
            search: '?q=a%26b%3Dc%23d',
        });
    });

    it('builds the image route with id and encoded path', () => {
        expect(getImageSearchRoute(9, '/pics/a&b#1.png')).toEqual({
            pathname: '/images',
            search: '?image=9&imagePath=%2Fpics%2Fa%26b%231.png',
        });
    });

    it('builds artist and album routes with encoded keys', () => {
        expect(getArtistSearchRoute('AC/DC & Co')).toEqual({
            pathname: '/music/artists',
            search: '?artist=AC%2FDC+%26+Co',
        });
        expect(getAlbumSearchRoute('x#y')).toEqual({
            pathname: '/music/albums',
            search: '?album=x%23y',
        });
    });

    it('builds the music playlist route by id', () => {
        expect(getPlaylistSearchRoute(buildVideoPlaylist({ scope: 'music', id: 3 }))).toEqual({
            pathname: '/music/playlists',
            search: '?playlist=3',
        });
    });

    it('builds the video playlist route carrying the playlist id', () => {
        expect(getPlaylistSearchRoute(buildVideoPlaylist({}))).toBe(
            '/videos/series/minha-serie?playlist=7'
        );
    });

    it('distinguishes video playlists with the same name by id', () => {
        const first = getPlaylistSearchRoute(buildVideoPlaylist({ id: 1 }));
        const second = getPlaylistSearchRoute(buildVideoPlaylist({ id: 2 }));
        expect(first).not.toBe(second);
    });

    it('falls back to the id slug when the name has no latin characters', () => {
        expect(getPlaylistSearchRoute(buildVideoPlaylist({ name: '日本', id: 12 }))).toBe(
            '/videos/series/12?playlist=12'
        );
    });

    it('routes folder-sourced playlists to the folders section', () => {
        expect(
            getPlaylistSearchRoute(buildVideoPlaylist({ source_path: '/v', description: 'folder' }))
        ).toBe('/videos/folders/minha-serie?playlist=7');
    });
});
