import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';

export function usePlaylistTrackHandlers() {
	const { addToQueue, getMusicArtist, getMusicTitle, musicMetadata } = useGlobalMusic();

	return {
		addToQueue,
		getMusicArtist,
		getMusicTitle,
		musicMetadata,
	};
}
