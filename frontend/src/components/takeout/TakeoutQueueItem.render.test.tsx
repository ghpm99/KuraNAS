import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import TakeoutQueueItem from './TakeoutQueueItem';
import { createQueueEntry, type TakeoutQueueEntry } from './takeoutUploadQueue';

const entryWith = (changes: Partial<TakeoutQueueEntry>): TakeoutQueueEntry => ({
    ...createQueueEntry(new File(['hello'], 'takeout.zip'), 'entry-1'),
    ...changes,
});

describe('TakeoutQueueItem without mocks', () => {
    it('shows the name and the waiting text for a queued entry', () => {
        const { container } = renderWithoutBackend(<TakeoutQueueItem entry={entryWith({})} />);

        expect(screen.getByText(/takeout\.zip/)).toBeInTheDocument();
        expect(screen.getByText('TAKEOUT_QUEUE_WAITING')).toBeInTheDocument();
        expect(container.textContent).not.toContain('undefined');
    });

    it('shows a determinate progress bar while uploading', () => {
        renderWithoutBackend(
            <TakeoutQueueItem entry={entryWith({ state: 'uploading', uploadedBytes: 2 })} />
        );

        expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '40');
    });

    it('shows the progress bar and the reconnecting text while waiting for the connection', () => {
        renderWithoutBackend(
            <TakeoutQueueItem entry={entryWith({ state: 'waitingConnection', uploadedBytes: 2 })} />
        );

        expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '40');
        expect(screen.getByText('TAKEOUT_UPLOAD_WAITING_CONNECTION')).toBeInTheDocument();
    });

    it('shows the processing text while completing', () => {
        renderWithoutBackend(<TakeoutQueueItem entry={entryWith({ state: 'completing' })} />);

        expect(screen.getByText('TAKEOUT_PROCESSING')).toBeInTheDocument();
    });

    it('shows success and the import progress for a done entry against an absent backend', () => {
        renderWithoutBackend(
            <TakeoutQueueItem entry={entryWith({ state: 'done', importId: 7 })} />
        );

        expect(screen.getByText('TAKEOUT_UPLOAD_COMPLETE')).toBeInTheDocument();
        expect(screen.getByText(/TAKEOUT_IMPORT_PROGRESS/)).toBeInTheDocument();
    });

    it('shows the error message for error and rejected entries', () => {
        const { rerender } = renderWithoutBackend(
            <TakeoutQueueItem
                entry={entryWith({ state: 'error', errorMessage: 'server said no' })}
            />
        );
        expect(screen.getByText('server said no')).toBeInTheDocument();

        rerender(
            <TakeoutQueueItem entry={entryWith({ state: 'rejected', errorMessage: 'only zip' })} />
        );
        expect(screen.getByText('only zip')).toBeInTheDocument();
    });
});
