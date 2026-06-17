import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { VideoContentProvider } from './videoContentProvider';

describe('VideoContentProvider (no-mock render)', () => {
    it('renders without throwing when the backend is unavailable', () => {
        expectRendersWithoutBackend(
            <VideoContentProvider>
                <div>videos</div>
            </VideoContentProvider>
        );
    });
});
