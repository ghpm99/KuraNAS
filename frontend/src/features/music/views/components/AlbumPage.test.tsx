import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import SettingsProvider from '@/components/providers/settingsProvider';
import { GlobalMusicProvider } from '@/features/music/providers/GlobalMusicProvider';
import AlbumPage from './AlbumPage';

describe('AlbumPage (no-mock render)', () => {
    it('renders without throwing when the backend is unavailable', () => {
        expectRendersWithoutBackend(
            <SettingsProvider>
                <GlobalMusicProvider>
                    <AlbumPage albumKey="band::album" onBack={jest.fn()} />
                </GlobalMusicProvider>
            </SettingsProvider>
        );
    });
});
