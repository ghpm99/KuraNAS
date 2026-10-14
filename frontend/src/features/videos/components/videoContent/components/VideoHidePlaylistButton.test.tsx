import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import VideoHidePlaylistButton from './VideoHidePlaylistButton';

describe('VideoHidePlaylistButton', () => {
    it('mounts without any backend and without a hide handler renders nothing', () => {
        const { container } = renderWithoutBackend(<VideoHidePlaylistButton isChanging={false} />);

        expect(container).toBeEmptyDOMElement();
    });

    it('calls the hide handler when clicked', () => {
        const onHide = jest.fn();
        renderWithoutBackend(<VideoHidePlaylistButton isChanging={false} onHide={onHide} />);

        fireEvent.click(screen.getByRole('button', { name: /VIDEO_HIDE_PLAYLIST/ }));

        expect(onHide).toHaveBeenCalledTimes(1);
    });

    it('is disabled while the visibility change is in flight', () => {
        renderWithoutBackend(<VideoHidePlaylistButton isChanging onHide={jest.fn()} />);

        expect(screen.getByRole('button', { name: /VIDEO_HIDE_PLAYLIST/ })).toBeDisabled();
    });
});
