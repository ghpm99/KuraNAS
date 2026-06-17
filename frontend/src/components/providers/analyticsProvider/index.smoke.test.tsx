import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { AnalyticsProvider } from './index';

describe('AnalyticsProvider (no-mock render)', () => {
    it('renders without throwing when the backend is unavailable', () => {
        expectRendersWithoutBackend(
            <AnalyticsProvider>
                <div>analytics</div>
            </AnalyticsProvider>
        );
    });
});
