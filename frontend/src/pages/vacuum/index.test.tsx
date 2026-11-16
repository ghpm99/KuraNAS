import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import VacuumPage from './index';

describe('VacuumPage', () => {
    it('mounts without a backend', () => {
        expect(() => renderWithoutBackend(<VacuumPage />)).not.toThrow();
    });
});
