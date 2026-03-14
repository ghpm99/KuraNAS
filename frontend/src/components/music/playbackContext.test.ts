import {
	createAlbumPlaybackContext,
	createAllTracksPlaybackContext,
	createArtistPlaybackContext,
	createFolderPlaybackContext,
	createGenrePlaybackContext,
	createPlaylistPlaybackContext,
	createRouteMusicPlaybackContext,
} from './playbackContext';

describe('music playback context helpers', () => {
	it('builds domain contexts with the expected routes and labels', () => {
		expect(createAllTracksPlaybackContext()).toMatchObject({
			labelKey: 'MUSIC_ALL_TRACKS',
			href: '/music',
		});
		expect(createArtistPlaybackContext('Artist A')).toMatchObject({
			labelKey: 'MUSIC_PLAYBACK_CONTEXT_ARTIST',
			labelParams: { name: 'Artist A' },
			href: '/music/artists',
		});
		expect(createAlbumPlaybackContext('Album A')).toMatchObject({
			labelKey: 'MUSIC_PLAYBACK_CONTEXT_ALBUM',
			labelParams: { name: 'Album A' },
			href: '/music/albums',
		});
		expect(createGenrePlaybackContext('Jazz')).toMatchObject({
			labelKey: 'MUSIC_PLAYBACK_CONTEXT_GENRE',
			labelParams: { name: 'Jazz' },
			href: '/music/genres',
		});
		expect(createFolderPlaybackContext('/library/jazz')).toMatchObject({
			labelKey: 'MUSIC_PLAYBACK_CONTEXT_FOLDER',
			labelParams: { name: '/library/jazz' },
			href: '/music/folders',
		});
		expect(createPlaylistPlaybackContext({ id: 7, name: 'Mix' })).toMatchObject({
			labelKey: 'MUSIC_PLAYBACK_CONTEXT_PLAYLIST',
			labelParams: { name: 'Mix' },
			href: '/music/playlists',
			playlistId: 7,
		});
	});

	it('maps current routes into playback contexts', () => {
		expect(createRouteMusicPlaybackContext('/files', '?path=%2Fmusic')).toMatchObject({
			labelKey: 'FILES',
			href: '/files?path=%2Fmusic',
		});
		expect(createRouteMusicPlaybackContext('/favorites')).toMatchObject({
			labelKey: 'STARRED_FILES',
			href: '/favorites',
		});
		expect(createRouteMusicPlaybackContext('/starred')).toMatchObject({
			labelKey: 'STARRED_FILES',
			href: '/starred',
		});
		expect(createRouteMusicPlaybackContext('/home')).toMatchObject({
			labelKey: 'HOME',
			href: '/home',
		});
		expect(createRouteMusicPlaybackContext('/music/albums')).toMatchObject({
			labelKey: 'MUSIC_ALL_TRACKS',
			href: '/music',
		});
		expect(createRouteMusicPlaybackContext('/analytics')).toMatchObject({
			labelKey: 'NAV_MUSIC',
			href: '/analytics',
		});
	});
});
