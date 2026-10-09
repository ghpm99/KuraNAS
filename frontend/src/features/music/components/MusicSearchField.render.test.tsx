import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import MusicSearchField from './MusicSearchField';

describe('MusicSearchField without a backend', () => {
    it('renders without service mocks', () => {
        expectRendersWithoutBackend(<MusicSearchField />);
    });
});
