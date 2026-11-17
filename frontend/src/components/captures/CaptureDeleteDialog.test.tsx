import { fireEvent, render, screen } from '@testing-library/react';
import type { Capture } from '@/types/captures';
import CaptureDeleteDialog from './CaptureDeleteDialog';

const receivedCapture: Capture = {
    id: 3,
    name: 'episodio.mp4',
    file_name: 'episodio.mp4',
    file_path: '/captures/episodio.mp4',
    media_type: 'video',
    mime_type: 'video/mp4',
    size: 1024,
    episode_key: 'serie-s01e01',
    created_at: '2026-11-16T10:00:00Z',
    status: 'uploaded',
};

const promotedCapture: Capture = { ...receivedCapture, status: 'promoted', file_id: 42 };

describe('CaptureDeleteDialog', () => {
    it('renders nothing visible without a capture and without any provider or service', () => {
        render(<CaptureDeleteDialog capture={null} onCancel={jest.fn()} onConfirm={jest.fn()} />);

        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('warns that the received file will be erased for a capture not yet promoted', () => {
        render(<CaptureDeleteDialog capture={receivedCapture} onCancel={jest.fn()} onConfirm={jest.fn()} />);

        expect(screen.getByText('CAPTURE_DELETE_UNPROMOTED_MESSAGE')).toBeInTheDocument();
        expect(screen.queryByText('CAPTURE_DELETE_PROMOTED_MESSAGE')).not.toBeInTheDocument();
    });

    it('says the library video is kept for a promoted capture', () => {
        render(<CaptureDeleteDialog capture={promotedCapture} onCancel={jest.fn()} onConfirm={jest.fn()} />);

        expect(screen.getByText('CAPTURE_DELETE_PROMOTED_MESSAGE')).toBeInTheDocument();
    });

    it('treats a capture with a library file id as promoted even before the status changes', () => {
        render(
            <CaptureDeleteDialog
                capture={{ ...receivedCapture, file_id: 42 }}
                onCancel={jest.fn()}
                onConfirm={jest.fn()}
            />
        );

        expect(screen.getByText('CAPTURE_DELETE_PROMOTED_MESSAGE')).toBeInTheDocument();
    });

    it('calls cancel and confirm from their buttons', () => {
        const onCancel = jest.fn();
        const onConfirm = jest.fn();
        render(<CaptureDeleteDialog capture={receivedCapture} onCancel={onCancel} onConfirm={onConfirm} />);

        fireEvent.click(screen.getByRole('button', { name: 'ACTION_CANCEL' }));
        fireEvent.click(screen.getByRole('button', { name: 'CAPTURES_DELETE' }));

        expect(onCancel).toHaveBeenCalledTimes(1);
        expect(onConfirm).toHaveBeenCalledTimes(1);
    });
});
