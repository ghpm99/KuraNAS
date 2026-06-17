import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { MusicProvider } from './musicProvider';

describe('MusicProvider (no-mock render)', () => {
    it('renders without throwing when the backend is unavailable', () => {
        expectRendersWithoutBackend(
            <MusicProvider>
                <div>music</div>
            </MusicProvider>
        );
    });
});
