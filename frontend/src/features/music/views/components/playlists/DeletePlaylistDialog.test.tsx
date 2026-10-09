import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DeletePlaylistDialog from './DeletePlaylistDialog';

describe('DeletePlaylistDialog', () => {
    it('renders without a pending playlist and without any backend', () => {
        expect(() =>
            renderWithoutBackend(
                <DeletePlaylistDialog playlistName={null} onConfirm={jest.fn()} onCancel={jest.fn()} />
            )
        ).not.toThrow();
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('asks for confirmation naming the playlist and reports the choice', () => {
        const onConfirm = jest.fn();
        const onCancel = jest.fn();
        renderWithoutBackend(
            <DeletePlaylistDialog playlistName="Mix" onConfirm={onConfirm} onCancel={onCancel} />
        );

        expect(screen.getByText('MUSIC_PLAYLIST_DELETE_CONFIRM_TITLE')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: 'ACTION_CANCEL' }));
        expect(onCancel).toHaveBeenCalled();
        expect(onConfirm).not.toHaveBeenCalled();

        fireEvent.click(screen.getByRole('button', { name: 'DELETE' }));
        expect(onConfirm).toHaveBeenCalled();
    });
});
