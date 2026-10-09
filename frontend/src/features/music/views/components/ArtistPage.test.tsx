import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import SettingsProvider from '@/components/providers/settingsProvider';
import { GlobalMusicProvider } from '@/features/music/providers/GlobalMusicProvider';
import ArtistPage from './ArtistPage';

describe('ArtistPage (no-mock render)', () => {
    it('renders without throwing when the backend is unavailable', () => {
        expectRendersWithoutBackend(
            <SettingsProvider>
                <GlobalMusicProvider>
                    <ArtistPage artistKey="band" onBack={jest.fn()} />
                </GlobalMusicProvider>
            </SettingsProvider>
        );
    });
});
