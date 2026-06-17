import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { AboutProvider } from './index';

describe('AboutProvider (no-mock render)', () => {
    it('renders without throwing when the backend is unavailable', () => {
        expectRendersWithoutBackend(
            <AboutProvider>
                <div>about</div>
            </AboutProvider>
        );
    });
});
