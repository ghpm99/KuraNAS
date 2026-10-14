import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import VideoNewPlaylistButton from './VideoNewPlaylistButton';

describe('VideoNewPlaylistButton', () => {
    it('mounts without any backend and opens no dialog until clicked', () => {
        renderWithoutBackend(<VideoNewPlaylistButton isCreating={false} onCreate={jest.fn()} />);

        expect(screen.getByRole('button', { name: /VIDEO_NEW_PLAYLIST/ })).toBeInTheDocument();
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('asks for a name in a dialog and creates the playlist with it', () => {
        const onCreate = jest.fn();
        renderWithoutBackend(<VideoNewPlaylistButton isCreating={false} onCreate={onCreate} />);

        fireEvent.click(screen.getByRole('button', { name: /VIDEO_NEW_PLAYLIST/ }));
        fireEvent.change(screen.getByLabelText('VIDEO_NEW_PLAYLIST_NAME'), {
            target: { value: 'Favoritos' },
        });
        fireEvent.click(screen.getByRole('button', { name: 'ACTION_CREATE' }));

        expect(onCreate).toHaveBeenCalledTimes(1);
        expect(onCreate).toHaveBeenCalledWith('Favoritos');
    });
});
