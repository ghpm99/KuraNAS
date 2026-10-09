import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import SettingsProvider from '@/components/providers/settingsProvider';
import { GlobalMusicProvider } from '@/features/music/providers/GlobalMusicProvider';
import AlbumsView from './AlbumsView';
import ArtistsView from './ArtistsView';
import FoldersView from './FoldersView';
import GenresView from './GenresView';

describe('music collection views (no-mock render)', () => {
    it.each([
        ['ArtistsView', <ArtistsView key="artists" />],
        ['AlbumsView', <AlbumsView key="albums" />],
        ['GenresView', <GenresView key="genres" />],
        ['FoldersView', <FoldersView key="folders" />],
    ])('%s renders without throwing when the backend is unavailable', (_name, view) => {
        expectRendersWithoutBackend(
            <SettingsProvider>
                <GlobalMusicProvider>{view}</GlobalMusicProvider>
            </SettingsProvider>
        );
    });
});
