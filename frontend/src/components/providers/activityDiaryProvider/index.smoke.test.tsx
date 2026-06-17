import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import ActivityDiaryProvider from './index';

describe('ActivityDiaryProvider (no-mock render)', () => {
    it('renders without throwing when the backend is unavailable', () => {
        expectRendersWithoutBackend(
            <ActivityDiaryProvider>
                <div>diary</div>
            </ActivityDiaryProvider>
        );
    });
});
