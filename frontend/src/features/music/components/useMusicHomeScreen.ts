import { getMusicRoute } from '@/app/routes';
import {
    createAlbumPlaybackContext,
    createArtistPlaybackContext,
    createPlaylistPlaybackContext,
} from '@/features/music/components/playbackContext';
import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';
import { getMusicTitle, getMusicArtist } from '@/utils/music';
import { getPlaylistQueue } from '@/service/playlist';
import { getMusicHomeCatalog, getMusicQueueByAlbum, getMusicQueueByArtist } from '@/service/music';
import { queueToTracks } from '@/features/music/components/musicQueueTracks';
import { useQuery } from '@tanstack/react-query';
import { useCallback, useState } from 'react';
import type { MusicAlbum, MusicArtist, MusicListSort } from '@/types/music';

const featuredPlaylistPageSize = 4;
const recentHighlightsSort: MusicListSort = { sort: 'recent', order: 'desc' };

const getActionKey = (type: 'playlist' | 'artist' | 'album', value: string | number) =>
    `${type}-${value}`;

export const useMusicHomeScreen = () => {
    const {
        currentIndex,
        currentTrack,
        hasQueue,
        playbackContext,
        queue,
        replaceQueue,
        toggleQueue,
    } = useGlobalMusic();
    const [pendingActionKey, setPendingActionKey] = useState<string | null>(null);

    const {
        data: homeCatalog,
        isLoading: isLoadingCatalog,
        status,
    } = useQuery({
        queryKey: ['music-home', 'catalog'],
        queryFn: () => getMusicHomeCatalog(featuredPlaylistPageSize, recentHighlightsSort),
    });

    const featuredPlaylists = homeCatalog?.playlists ?? [];
    const artistHighlights = homeCatalog?.artists ?? [];
    const albumHighlights = homeCatalog?.albums ?? [];
    const totalTracks = homeCatalog?.summary.total_tracks ?? 0;
    const totalArtists = homeCatalog?.summary.total_artists ?? 0;
    const totalAlbums = homeCatalog?.summary.total_albums ?? 0;

    const playPlaylist = useCallback(
        async (playlistId: number, playlistName: string) => {
            const actionKey = getActionKey('playlist', playlistId);
            setPendingActionKey(actionKey);

            try {
                const playlistTracks = queueToTracks(await getPlaylistQueue(playlistId));

                if (playlistTracks.length > 0) {
                    replaceQueue(
                        playlistTracks,
                        0,
                        createPlaylistPlaybackContext({
                            id: playlistId,
                            name: playlistName,
                        })
                    );
                }
            } finally {
                setPendingActionKey((currentKey) => (currentKey === actionKey ? null : currentKey));
            }
        },
        [replaceQueue]
    );

    const playArtist = useCallback(
        async (artist: MusicArtist) => {
            const actionKey = getActionKey('artist', artist.key);
            setPendingActionKey(actionKey);

            try {
                const artistTracks = queueToTracks(await getMusicQueueByArtist(artist.key));

                if (artistTracks.length > 0) {
                    replaceQueue(artistTracks, 0, createArtistPlaybackContext(artist.artist));
                }
            } finally {
                setPendingActionKey((currentKey) => (currentKey === actionKey ? null : currentKey));
            }
        },
        [replaceQueue]
    );

    const playAlbum = useCallback(
        async (album: MusicAlbum) => {
            const actionKey = getActionKey('album', album.key);
            setPendingActionKey(actionKey);

            try {
                const albumTracks = queueToTracks(await getMusicQueueByAlbum(album.key));

                if (albumTracks.length > 0) {
                    replaceQueue(albumTracks, 0, createAlbumPlaybackContext(album.album));
                }
            } finally {
                setPendingActionKey((currentKey) => (currentKey === actionKey ? null : currentKey));
            }
        },
        [replaceQueue]
    );

    const currentTrackTitle = currentTrack ? getMusicTitle(currentTrack) : '';
    const currentTrackArtist = currentTrack ? getMusicArtist(currentTrack) : '';
    const nextTracks = queue
        .filter((_, index) => currentIndex !== undefined && index > currentIndex)
        .slice(0, 3)
        .map((track) => ({
            id: track.id,
            title: getMusicTitle(track),
            artist: getMusicArtist(track),
        }));

    return {
        status,
        hasQueue,
        queueCount: queue.length,
        totalTracks,
        totalArtists,
        totalAlbums,
        totalPlaylists: featuredPlaylists.length,
        currentTrackTitle,
        currentTrackArtist,
        nextTracks,
        playbackContext,
        returnToContextHref: playbackContext?.href,
        openQueue: toggleQueue,
        isLoadingPlaylists: isLoadingCatalog,
        featuredPlaylists: featuredPlaylists.map((playlist) => ({
            ...playlist,
            href: getMusicRoute('playlists'),
            actionKey: getActionKey('playlist', playlist.id),
        })),
        artistHighlights: artistHighlights.map((artist) => ({
            ...artist,
            href: getMusicRoute('artists'),
            actionKey: getActionKey('artist', artist.key),
        })),
        albumHighlights: albumHighlights.map((album) => ({
            ...album,
            href: getMusicRoute('albums'),
            actionKey: getActionKey('album', album.key),
        })),
        isActionPending: (actionKey: string) => pendingActionKey === actionKey,
        playPlaylist,
        playArtist,
        playAlbum,
    };
};

export type MusicHomeScreenData = ReturnType<typeof useMusicHomeScreen>;
export type MusicHomeArtistCard = MusicArtist & {
    href: string;
    actionKey: string;
};
export type MusicHomeAlbumCard = MusicAlbum & {
    href: string;
    actionKey: string;
};
