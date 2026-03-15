import { getMusicTitle, getMusicArtist, musicMetadata } from '@/utils/music';

export function usePlaylistTrackHandlers() {
	return {
		getMusicArtist,
		getMusicTitle,
		musicMetadata,
	};
}
