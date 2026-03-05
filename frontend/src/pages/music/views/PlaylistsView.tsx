import { PlaylistsProvider } from '@/components/providers/playlistsProvider';
import PlaylistsScreen from './components/playlists/PlaylistsScreen';

export default function PlaylistsView() {
	return (
		<PlaylistsProvider>
			<PlaylistsScreen />
		</PlaylistsProvider>
	);
}
