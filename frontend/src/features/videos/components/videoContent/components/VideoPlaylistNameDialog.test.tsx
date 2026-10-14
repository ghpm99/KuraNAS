import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import VideoPlaylistNameDialog from './VideoPlaylistNameDialog';

const renderDialog = (overrides: Partial<Parameters<typeof VideoPlaylistNameDialog>[0]> = {}) => {
    const onSubmit = jest.fn();
    const onClose = jest.fn();
    renderWithoutBackend(
        <VideoPlaylistNameDialog
            isOpen
            isSubmitting={false}
            confirmLabelKey="ACTION_CREATE"
            onClose={onClose}
            onSubmit={onSubmit}
            {...overrides}
        />
    );
    return { onSubmit, onClose };
};

describe('VideoPlaylistNameDialog', () => {
    it('mounts without any backend and keeps submit disabled while the name is empty', () => {
        renderDialog();

        expect(screen.getByRole('button', { name: 'ACTION_CREATE' })).toBeDisabled();
    });

    it('submits the trimmed name', () => {
        const { onSubmit } = renderDialog();

        fireEvent.change(screen.getByLabelText('VIDEO_NEW_PLAYLIST_NAME'), {
            target: { value: '  Favoritos  ' },
        });
        fireEvent.click(screen.getByRole('button', { name: 'ACTION_CREATE' }));

        expect(onSubmit).toHaveBeenCalledWith('Favoritos');
    });

    it('submits with Enter and ignores blank names', () => {
        const { onSubmit } = renderDialog();
        const nameField = screen.getByLabelText('VIDEO_NEW_PLAYLIST_NAME');

        fireEvent.change(nameField, { target: { value: '   ' } });
        fireEvent.keyDown(nameField, { key: 'Enter' });
        expect(onSubmit).not.toHaveBeenCalled();

        fireEvent.change(nameField, { target: { value: 'Lista' } });
        fireEvent.keyDown(nameField, { key: 'Enter' });
        expect(onSubmit).toHaveBeenCalledWith('Lista');
    });

    it('cancels without submitting', () => {
        const { onSubmit, onClose } = renderDialog();

        fireEvent.click(screen.getByRole('button', { name: 'ACTION_CANCEL' }));

        expect(onClose).toHaveBeenCalled();
        expect(onSubmit).not.toHaveBeenCalled();
    });

    it('blocks submission while a creation is in flight', () => {
        const { onSubmit } = renderDialog({ isSubmitting: true });

        fireEvent.change(screen.getByLabelText('VIDEO_NEW_PLAYLIST_NAME'), {
            target: { value: 'Lista' },
        });

        expect(screen.queryByRole('button', { name: 'ACTION_CREATE' })).not.toBeInTheDocument();
        expect(onSubmit).not.toHaveBeenCalled();
    });
});
