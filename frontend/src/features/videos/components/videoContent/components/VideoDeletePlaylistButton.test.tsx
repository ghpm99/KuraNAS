import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import VideoDeletePlaylistButton from './VideoDeletePlaylistButton';

describe('VideoDeletePlaylistButton', () => {
    it('mounts without any backend and without a delete handler renders nothing', () => {
        const { container } = renderWithoutBackend(
            <VideoDeletePlaylistButton
                playlist={{ name: 'x', is_auto: false }}
                isDeleting={false}
            />
        );

        expect(container).toBeEmptyDOMElement();
    });

    it('is absent for automatic playlists', () => {
        renderWithoutBackend(
            <VideoDeletePlaylistButton
                playlist={{ name: 'Auto', is_auto: true }}
                isDeleting={false}
                onDelete={jest.fn()}
            />
        );

        expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });

    it('is absent when the backend omits is_auto', () => {
        renderWithoutBackend(
            <VideoDeletePlaylistButton
                playlist={{ name: 'x' } as { name: string; is_auto: boolean }}
                isDeleting={false}
                onDelete={jest.fn()}
            />
        );

        expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });

    it('asks for confirmation before deleting a custom playlist', () => {
        const onDelete = jest.fn();
        renderWithoutBackend(
            <VideoDeletePlaylistButton
                playlist={{ name: 'Favoritos', is_auto: false }}
                isDeleting={false}
                onDelete={onDelete}
            />
        );

        fireEvent.click(screen.getByRole('button', { name: /VIDEO_DELETE_PLAYLIST/ }));
        expect(onDelete).not.toHaveBeenCalled();
        expect(screen.getByText('VIDEO_DELETE_PLAYLIST_CONFIRM_TITLE')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'VIDEO_DELETE_PLAYLIST' }));

        expect(onDelete).toHaveBeenCalledTimes(1);
    });

    it('does not delete when the confirmation is cancelled', () => {
        const onDelete = jest.fn();
        renderWithoutBackend(
            <VideoDeletePlaylistButton
                playlist={{ name: 'Favoritos', is_auto: false }}
                isDeleting={false}
                onDelete={onDelete}
            />
        );

        fireEvent.click(screen.getByRole('button', { name: /VIDEO_DELETE_PLAYLIST/ }));
        fireEvent.click(screen.getByRole('button', { name: 'ACTION_CANCEL' }));

        expect(onDelete).not.toHaveBeenCalled();
    });
});
