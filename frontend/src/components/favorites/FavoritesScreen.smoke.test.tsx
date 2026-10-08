import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import FileProvider from '@/features/files/providers/fileProvider';
import { GlobalMusicProvider } from '@/features/music/providers/GlobalMusicProvider';
import SettingsProvider from '@/components/providers/settingsProvider';
import FavoritesScreen from './FavoritesScreen';

describe('FavoritesScreen (no-mock render)', () => {
    it('renders without throwing when the backend is unavailable', () => {
        expectRendersWithoutBackend(
            <SettingsProvider>
                <GlobalMusicProvider>
                    <FileProvider>
                        <FavoritesScreen />
                    </FileProvider>
                </GlobalMusicProvider>
            </SettingsProvider>
        );
    });
});
