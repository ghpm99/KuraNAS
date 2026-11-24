import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { PlaylistRules } from '@/types/playlist';
import PlaylistRuleCriteriaSummary from './PlaylistRuleCriteriaSummary';
import { formatAlbumKeyLabel } from './formatAlbumKeyLabel';

const rules: PlaylistRules = {
    match: 'all',
    genres: ['rock'],
    artists: ['queen'],
    albums: ['queen::a night at the opera'],
    folders: ['/music/rock'],
    year_from: 1970,
    year_to: null,
    favorites_only: true,
};

describe('PlaylistRuleCriteriaSummary', () => {
    it('renders without a backend and without rules', () => {
        expect(() =>
            renderWithoutBackend(<PlaylistRuleCriteriaSummary rules={undefined} />)
        ).not.toThrow();
        expect(() =>
            renderWithoutBackend(
                <PlaylistRuleCriteriaSummary rules={{ match: 'any' } as unknown as PlaylistRules} />
            )
        ).not.toThrow();
    });

    it('shows every criterion with the album label instead of the raw key', () => {
        renderWithoutBackend(<PlaylistRuleCriteriaSummary rules={rules} />);

        expect(screen.getByText('rock')).toBeInTheDocument();
        expect(screen.getByText('queen')).toBeInTheDocument();
        expect(screen.getByText('a night at the opera - queen')).toBeInTheDocument();
        expect(screen.queryByText('queen::a night at the opera')).not.toBeInTheDocument();
        expect(screen.getByText('/music/rock')).toBeInTheDocument();
        expect(screen.getByTestId('playlist-rule-criteria')).toBeInTheDocument();
    });

    it('formats album keys with or without artist part', () => {
        expect(formatAlbumKeyLabel('queen::news')).toBe('news - queen');
        expect(formatAlbumKeyLabel('::news')).toBe('news');
        expect(formatAlbumKeyLabel('news')).toBe('news');
    });
});
