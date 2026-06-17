import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { PlaylistsProvider } from './playlistsProvider';

describe('PlaylistsProvider (no-mock render)', () => {
    it('renders without throwing when the backend is unavailable', () => {
        expectRendersWithoutBackend(
            <PlaylistsProvider>
                <div>playlists</div>
            </PlaylistsProvider>
        );
    });
});
