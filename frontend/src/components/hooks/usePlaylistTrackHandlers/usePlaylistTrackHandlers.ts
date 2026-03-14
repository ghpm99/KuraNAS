import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';

export function usePlaylistTrackHandlers() {
	const { getMusicArtist, getMusicTitle, musicMetadata } = useGlobalMusic();

	return {
		getMusicArtist,
		getMusicTitle,
		musicMetadata,
	};
}
