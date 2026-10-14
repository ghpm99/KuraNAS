import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import VideoFeedbackSnackbar from './VideoFeedbackSnackbar';

describe('VideoFeedbackSnackbar', () => {
    it('mounts without any backend and shows the message without an action', () => {
        renderWithoutBackend(
            <VideoFeedbackSnackbar open message="Pronto" severity="success" onClose={jest.fn()} />
        );

        expect(screen.getByText('Pronto')).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'Desfazer' })).not.toBeInTheDocument();
    });

    it('runs the action and closes when the action button is clicked', () => {
        const onClick = jest.fn();
        const onClose = jest.fn();
        renderWithoutBackend(
            <VideoFeedbackSnackbar
                open
                message="Oculta"
                severity="success"
                action={{ label: 'Desfazer', onClick }}
                onClose={onClose}
            />
        );

        fireEvent.click(screen.getByRole('button', { name: 'Desfazer' }));

        expect(onClick).toHaveBeenCalledTimes(1);
        expect(onClose).toHaveBeenCalledTimes(1);
    });
});
