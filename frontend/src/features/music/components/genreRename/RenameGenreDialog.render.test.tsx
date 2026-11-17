import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import RenameGenreDialog from './RenameGenreDialog';

describe('RenameGenreDialog with no service mock and no backend', () => {
    it('mounts with the current genre name and review enabled', () => {
        renderWithoutBackend(
            <RenameGenreDialog
                genreKey="rock"
                genreName="Rock"
                trackCount={12}
                onRenamed={jest.fn()}
                onClose={jest.fn()}
            />
        );

        expect(screen.getByLabelText('MUSIC_GENRE_RENAME_FIELD_NAME')).toHaveValue('Rock');
        expect(screen.getByRole('button', { name: 'MUSIC_GENRE_RENAME_REVIEW' })).toBeEnabled();
    });
});
