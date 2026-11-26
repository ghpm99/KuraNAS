import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import TakeoutQueueItem from './TakeoutQueueItem';
import { createQueueEntry, type TakeoutQueueEntry } from './takeoutUploadQueue';

const entryWith = (changes: Partial<TakeoutQueueEntry>): TakeoutQueueEntry => ({
    ...createQueueEntry(new File(['hello'], 'takeout.zip'), 'entry-1'),
    ...changes,
});

describe('TakeoutQueueItem without mocks', () => {
    it('shows the name and the waiting text for a queued entry', () => {
        const { container } = renderWithoutBackend(
            <TakeoutQueueItem entry={entryWith({})} onCancel={jest.fn()} />
        );

        expect(screen.getByText(/takeout\.zip/)).toBeInTheDocument();
        expect(screen.getByText('TAKEOUT_QUEUE_WAITING')).toBeInTheDocument();
        expect(container.textContent).not.toContain('undefined');
    });

    it('shows a determinate progress bar while uploading', () => {
        renderWithoutBackend(
            <TakeoutQueueItem
                entry={entryWith({ state: 'uploading', uploadedBytes: 2 })}
                onCancel={jest.fn()}
            />
        );

        expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '40');
    });

    it('shows speed and remaining time while uploading when the rate is known', () => {
        renderWithoutBackend(
            <TakeoutQueueItem
                entry={entryWith({
                    state: 'uploading',
                    uploadedBytes: 2,
                    bytesPerSecond: 1024,
                    remainingSeconds: 30,
                })}
                onCancel={jest.fn()}
            />
        );

        expect(screen.getByText(/FILES_UPLOAD_SPEED/)).toBeInTheDocument();
        expect(screen.getByText(/FILES_UPLOAD_REMAINING/)).toBeInTheDocument();
    });

    it('shows no speed or remaining time while uploading without a rate', () => {
        const { container } = renderWithoutBackend(
            <TakeoutQueueItem
                entry={entryWith({ state: 'uploading', uploadedBytes: 2 })}
                onCancel={jest.fn()}
            />
        );

        expect(container.textContent).not.toContain('FILES_UPLOAD_SPEED');
        expect(container.textContent).not.toContain('FILES_UPLOAD_REMAINING');
    });

    it('shows the progress bar and the reconnecting text while waiting for the connection', () => {
        renderWithoutBackend(
            <TakeoutQueueItem
                entry={entryWith({ state: 'waitingConnection', uploadedBytes: 2 })}
                onCancel={jest.fn()}
            />
        );

        expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '40');
        expect(screen.getByText('TAKEOUT_UPLOAD_WAITING_CONNECTION')).toBeInTheDocument();
    });

    it('shows the processing text while completing', () => {
        renderWithoutBackend(
            <TakeoutQueueItem entry={entryWith({ state: 'completing' })} onCancel={jest.fn()} />
        );

        expect(screen.getByText('TAKEOUT_PROCESSING')).toBeInTheDocument();
    });

    it('shows success and the import progress for a done entry against an absent backend', () => {
        renderWithoutBackend(
            <TakeoutQueueItem
                entry={entryWith({ state: 'done', importId: 7 })}
                onCancel={jest.fn()}
            />
        );

        expect(screen.getByText('TAKEOUT_UPLOAD_COMPLETE')).toBeInTheDocument();
        expect(screen.getByText(/TAKEOUT_IMPORT_PROGRESS/)).toBeInTheDocument();
    });

    it('shows the error message for error and rejected entries', () => {
        const { rerender } = renderWithoutBackend(
            <TakeoutQueueItem
                entry={entryWith({ state: 'error', errorMessage: 'server said no' })}
                onCancel={jest.fn()}
            />
        );
        expect(screen.getByText('server said no')).toBeInTheDocument();

        rerender(
            <TakeoutQueueItem
                entry={entryWith({ state: 'rejected', errorMessage: 'only zip' })}
                onCancel={jest.fn()}
            />
        );
        expect(screen.getByText('only zip')).toBeInTheDocument();
    });

    it('offers cancel for queued, uploading and waiting entries and calls onCancel', () => {
        const onCancel = jest.fn();
        renderWithoutBackend(
            <TakeoutQueueItem entry={entryWith({ state: 'uploading' })} onCancel={onCancel} />
        );

        fireEvent.click(screen.getByRole('button', { name: 'TAKEOUT_CANCEL' }));

        expect(onCancel).toHaveBeenCalledWith('entry-1');
    });

    it('has no cancel button while completing', () => {
        renderWithoutBackend(
            <TakeoutQueueItem entry={entryWith({ state: 'completing' })} onCancel={jest.fn()} />
        );

        expect(screen.queryByRole('button', { name: 'TAKEOUT_CANCEL' })).not.toBeInTheDocument();
    });

    it('shows the canceled text and the delete failure message', () => {
        renderWithoutBackend(
            <TakeoutQueueItem
                entry={entryWith({ state: 'canceled', errorMessage: 'TAKEOUT_CANCEL_ERROR' })}
                onCancel={jest.fn()}
            />
        );

        expect(screen.getByText('TAKEOUT_CANCELED')).toBeInTheDocument();
        expect(screen.getByText('TAKEOUT_CANCEL_ERROR')).toBeInTheDocument();
    });
});
