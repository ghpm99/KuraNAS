import { getMusicRoute } from '@/app/routes';
import {
	buildMusicAlbumHighlights,
	buildMusicArtistHighlights,
	type MusicAlbumHighlight,
	type MusicArtistHighlight,
} from '@/components/music/musicHomeData';
import {
	createAlbumPlaybackContext,
	createArtistPlaybackContext,
	createPlaylistPlaybackContext,
} from '@/components/music/playbackContext';
import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';
import { useMusic } from '@/components/providers/musicProvider/musicProvider';
import { getPlaylistTracks, getPlaylists } from '@/service/playlist';
import { getMusicByAlbum, getMusicByArtist } from '@/service/music';
import { useQuery } from '@tanstack/react-query';
import { useCallback, useMemo, useState } from 'react';

const featuredPlaylistPageSize = 4;
const homeHighlightLimit = 4;

const getActionKey = (type: 'playlist' | 'artist' | 'album', value: string | number) => `${type}-${value}`;

export const useMusicHomeScreen = () => {
	const { status, music } = useMusic();
	const {
		currentIndex,
		currentTrack,
		getMusicArtist,
		getMusicTitle,
		hasQueue,
		playbackContext,
		queue,
		replaceQueue,
		toggleQueue,
	} = useGlobalMusic();
	const [pendingActionKey, setPendingActionKey] = useState<string | null>(null);

	const { data: playlistResponse, isLoading: isLoadingPlaylists } = useQuery({
		queryKey: ['music-home', 'playlists'],
		queryFn: () => getPlaylists(1, featuredPlaylistPageSize),
	});

	const featuredPlaylists = playlistResponse?.items ?? [];
	const artistHighlights = useMemo(() => buildMusicArtistHighlights(music, homeHighlightLimit), [music]);
	const albumHighlights = useMemo(() => buildMusicAlbumHighlights(music, homeHighlightLimit), [music]);
	const totalArtists = useMemo(() => buildMusicArtistHighlights(music, music.length).length, [music]);
	const totalAlbums = useMemo(() => buildMusicAlbumHighlights(music, music.length).length, [music]);

	const playPlaylist = useCallback(
		async (playlistId: number, playlistName: string) => {
			const actionKey = getActionKey('playlist', playlistId);
			setPendingActionKey(actionKey);

			try {
				const response = await getPlaylistTracks(playlistId, 1, 200);
				const playlistTracks = response.items.map((item) => item.file);

				if (playlistTracks.length > 0) {
					replaceQueue(playlistTracks, 0, createPlaylistPlaybackContext({ id: playlistId, name: playlistName }));
				}
			} finally {
				setPendingActionKey((currentKey) => (currentKey === actionKey ? null : currentKey));
			}
		},
		[replaceQueue],
	);

	const playArtist = useCallback(
		async (artist: string) => {
			const actionKey = getActionKey('artist', artist);
			setPendingActionKey(actionKey);

			try {
				const response = await getMusicByArtist(artist, 1, 200);

				if (response.items.length > 0) {
					replaceQueue(response.items, 0, createArtistPlaybackContext(artist));
				}
			} finally {
				setPendingActionKey((currentKey) => (currentKey === actionKey ? null : currentKey));
			}
		},
		[replaceQueue],
	);

	const playAlbum = useCallback(
		async (album: string) => {
			const actionKey = getActionKey('album', album);
			setPendingActionKey(actionKey);

			try {
				const response = await getMusicByAlbum(album, 1, 200);

				if (response.items.length > 0) {
					replaceQueue(response.items, 0, createAlbumPlaybackContext(album));
				}
			} finally {
				setPendingActionKey((currentKey) => (currentKey === actionKey ? null : currentKey));
			}
		},
		[replaceQueue],
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
		totalTracks: music.length,
		totalArtists,
		totalAlbums,
		totalPlaylists: featuredPlaylists.length,
		currentTrackTitle,
		currentTrackArtist,
		nextTracks,
		playbackContext,
		returnToContextHref: playbackContext?.href,
		openQueue: toggleQueue,
		isLoadingPlaylists,
		featuredPlaylists: featuredPlaylists.map((playlist) => ({
			...playlist,
			href: getMusicRoute('playlists'),
			actionKey: getActionKey('playlist', playlist.id),
		})),
		artistHighlights: artistHighlights.map((artist) => ({
			...artist,
			href: getMusicRoute('artists'),
			actionKey: getActionKey('artist', artist.artist),
		})),
		albumHighlights: albumHighlights.map((album) => ({
			...album,
			href: getMusicRoute('albums'),
			actionKey: getActionKey('album', album.album),
		})),
		isActionPending: (actionKey: string) => pendingActionKey === actionKey,
		playPlaylist,
		playArtist,
		playAlbum,
	};
};

export type MusicHomeScreenData = ReturnType<typeof useMusicHomeScreen>;
export type MusicHomeArtistCard = MusicArtistHighlight & { href: string; actionKey: string };
export type MusicHomeAlbumCard = MusicAlbumHighlight & { href: string; actionKey: string };
