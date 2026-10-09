import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import SearchView from './SearchView';

describe('SearchView (no-mock render)', () => {
    it('renders without throwing when the backend is unavailable', () => {
        expectRendersWithoutBackend(<SearchView />);
    });
});
