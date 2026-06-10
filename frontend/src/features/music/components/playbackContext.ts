import { appRoutes, getMusicRoute } from '@/app/routes';
import type { Playlist } from '@/types/playlist';

export interface MusicPlaybackContext {
    kind: 'route' | 'all-tracks' | 'playlist' | 'artist' | 'album' | 'genre' | 'folder';
    labelKey: string;
    labelParams?: Record<string, string>;
    href: string;
    playlistId?: number | null;
}

const buildRouteHref = (pathname: string, search = '') => `${pathname}${search}`;

export const createAllTracksPlaybackContext = (): MusicPlaybackContext => ({
    kind: 'all-tracks',
    labelKey: 'MUSIC_ALL_TRACKS',
    href: getMusicRoute('home'),
});

export const createPlaylistPlaybackContext = (
    playlist: Pick<Playlist, 'id' | 'name'>
): MusicPlaybackContext => ({
    kind: 'playlist',
    labelKey: 'MUSIC_PLAYBACK_CONTEXT_PLAYLIST',
    labelParams: { name: playlist.name },
    href: getMusicRoute('playlists'),
    playlistId: playlist.id,
});

export const createArtistPlaybackContext = (artist: string): MusicPlaybackContext => ({
    kind: 'artist',
    labelKey: 'MUSIC_PLAYBACK_CONTEXT_ARTIST',
    labelParams: { name: artist },
    href: getMusicRoute('artists'),
});

export const createAlbumPlaybackContext = (album: string): MusicPlaybackContext => ({
    kind: 'album',
    labelKey: 'MUSIC_PLAYBACK_CONTEXT_ALBUM',
    labelParams: { name: album },
    href: getMusicRoute('albums'),
});

export const createGenrePlaybackContext = (genre: string): MusicPlaybackContext => ({
    kind: 'genre',
    labelKey: 'MUSIC_PLAYBACK_CONTEXT_GENRE',
    labelParams: { name: genre },
    href: getMusicRoute('genres'),
});

export const createFolderPlaybackContext = (folder: string): MusicPlaybackContext => ({
    kind: 'folder',
    labelKey: 'MUSIC_PLAYBACK_CONTEXT_FOLDER',
    labelParams: { name: folder },
    href: getMusicRoute('folders'),
});

export const createRouteMusicPlaybackContext = (
    pathname: string,
    search = ''
): MusicPlaybackContext => {
    if (pathname.startsWith(appRoutes.files)) {
        return {
            kind: 'route',
            labelKey: 'FILES',
            href: buildRouteHref(pathname, search),
        };
    }

    if (pathname === appRoutes.favorites || pathname === appRoutes.legacyFavorites) {
        return {
            kind: 'route',
            labelKey: 'STARRED_FILES',
            href: buildRouteHref(pathname, search),
        };
    }

    if (pathname === appRoutes.home) {
        return {
            kind: 'route',
            labelKey: 'HOME',
            href: buildRouteHref(pathname, search),
        };
    }

    if (pathname.startsWith(appRoutes.music)) {
        return createAllTracksPlaybackContext();
    }

    return {
        kind: 'route',
        labelKey: 'NAV_MUSIC',
        href: buildRouteHref(pathname, search),
    };
};
