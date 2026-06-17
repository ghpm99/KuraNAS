import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { UIProvider } from './index';

describe('UIProvider (no-mock render)', () => {
    it('renders without throwing when the backend is unavailable', () => {
        expectRendersWithoutBackend(
            <UIProvider>
                <div>ui</div>
            </UIProvider>
        );
    });
});
