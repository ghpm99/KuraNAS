import { fireEvent, screen } from '@testing-library/react';
import { useLocation } from 'react-router-dom';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import CreateRulePlaylistFromGenreButton from './CreateRulePlaylistFromGenreButton';

const LocationProbe = () => {
    const location = useLocation();
    return <span data-testid="location">{`${location.pathname}${location.search}`}</span>;
};

describe('CreateRulePlaylistFromGenreButton', () => {
    it('renders without a backend', () => {
        expect(() =>
            renderWithoutBackend(<CreateRulePlaylistFromGenreButton genreKey="rock" />)
        ).not.toThrow();
    });

    it('navigates to the playlists route asking for a rule playlist of the genre', () => {
        renderWithoutBackend(
            <>
                <CreateRulePlaylistFromGenreButton genreKey="hip hop" />
                <LocationProbe />
            </>
        );

        fireEvent.click(screen.getByRole('button'));

        expect(screen.getByTestId('location').textContent).toMatch(
            /\/playlists\?create_rule_genre=hip%20hop$/
        );
    });
});
