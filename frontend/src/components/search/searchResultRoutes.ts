import { appRoutes, buildFilesUrl, getMusicRoute, getVideoRoute } from '@/app/routes';
import {
    getVideoDetailRoute,
    getVideoSectionForPlaylist,
} from '@/features/videos/components/navigation';
import type { GlobalSearchPlaylistResult } from '@/service/search';

type SearchRoute = { pathname: string; search: string };

const buildQueryString = (params: Record<string, string>) =>
    `?${new URLSearchParams(params).toString()}`;

const slugifyPlaylistName = (playlistName: string) =>
    playlistName
        .normalize('NFD')
        .replace(/[̀-ͯ]/g, '')
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/^-+|-+$/g, '');

export const getFileSearchRoute = (filePath: string) => buildFilesUrl(filePath);

export const getFilesQuerySearchRoute = (searchText: string): SearchRoute => ({
    pathname: appRoutes.files,
    search: buildQueryString({ q: searchText }),
});

export const getImagesQuerySearchRoute = (searchText: string): SearchRoute => ({
    pathname: appRoutes.images,
    search: buildQueryString({ q: searchText }),
});

export const getVideosQuerySearchRoute = (searchText: string): SearchRoute => ({
    pathname: getVideoRoute('folders'),
    search: buildQueryString({ q: searchText }),
});

export const getMusicQuerySearchRoute = (searchText: string): SearchRoute => ({
    pathname: getMusicRoute('search'),
    search: buildQueryString({ q: searchText }),
});

export const getImageSearchRoute = (imageId: number, imagePath: string): SearchRoute => ({
    pathname: appRoutes.images,
    search: buildQueryString({ image: String(imageId), imagePath }),
});

export const getArtistSearchRoute = (artistKey: string): SearchRoute => ({
    pathname: getMusicRoute('artists'),
    search: buildQueryString({ artist: artistKey }),
});

export const getAlbumSearchRoute = (albumKey: string): SearchRoute => ({
    pathname: getMusicRoute('albums'),
    search: buildQueryString({ album: albumKey }),
});

export const getPlaylistSearchRoute = (
    playlist: GlobalSearchPlaylistResult
): SearchRoute | string => {
    if (playlist.scope === 'music') {
        return {
            pathname: getMusicRoute('playlists'),
            search: buildQueryString({ playlist: String(playlist.id) }),
        };
    }

    const section = getVideoSectionForPlaylist({
        type: playlist.source_path ? playlist.description : 'custom',
        classification: playlist.classification,
    });
    return getVideoDetailRoute(
        section,
        slugifyPlaylistName(playlist.name) || String(playlist.id),
        playlist.id
    );
};
