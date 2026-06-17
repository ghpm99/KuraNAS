import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import FileProvider from './index';

describe('FileProvider (no-mock render)', () => {
    it('renders without throwing when the backend is unavailable', () => {
        expectRendersWithoutBackend(
            <FileProvider>
                <div>files</div>
            </FileProvider>
        );
    });
});
