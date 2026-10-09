import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import AddToPlaylistMenu from './AddToPlaylistMenu';

describe('AddToPlaylistMenu without mocks', () => {
    it('mounts open against an absent backend and still offers creating a playlist', async () => {
        renderWithoutBackend(
            <AddToPlaylistMenu fileId={1} anchorEl={document.body} onClose={jest.fn()} />
        );

        expect(await screen.findByText('MUSIC_NEW_PLAYLIST')).toBeInTheDocument();
    });
});
