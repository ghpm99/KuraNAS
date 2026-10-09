import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import MusicPlayHistorySections from './MusicPlayHistorySections';

describe('MusicPlayHistorySections without mocks', () => {
    it('mounts against an absent backend and shows both empty states', async () => {
        renderWithoutBackend(<MusicPlayHistorySections />);

        expect(await screen.findByText('MUSIC_HOME_MOST_PLAYED_EMPTY')).toBeInTheDocument();
        expect(screen.getByText('MUSIC_HOME_RECENT_PLAYS_EMPTY')).toBeInTheDocument();
    });
});
