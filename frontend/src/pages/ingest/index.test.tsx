import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import IngestPage from './index';

describe('IngestPage', () => {
    it('mounts without a backend', () => {
        expect(() => renderWithoutBackend(<IngestPage />)).not.toThrow();
    });
});
