import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { ImageProvider } from './imageProvider';

describe('ImageProvider (no-mock render)', () => {
    it('renders without throwing when the backend is unavailable', () => {
        expectRendersWithoutBackend(
            <ImageProvider>
                <div>image</div>
            </ImageProvider>
        );
    });
});
