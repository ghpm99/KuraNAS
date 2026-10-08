import { fireEvent, render, screen } from '@testing-library/react';
import UploadQueuePanel from './uploadQueuePanel';
import { UploadQueueContextProvider } from './uploadQueueContext';
import type { UploadItem, UploadQueue } from './uploadQueueTypes';

const buildItem = (overrides: Partial<UploadItem>): UploadItem => ({
    id: 'upload-1',
    file: new File(['x'], 'a.txt'),
    displayName: 'a.txt',
    onConflict: 'rename',
    status: 'queued',
    progress: 0,
    ...overrides,
});

const buildQueue = (items: UploadItem[]): UploadQueue => ({
    items,
    conflictPolicy: 'rename',
    setConflictPolicy: jest.fn(),
    enqueue: jest.fn(),
    cancel: jest.fn(),
    cancelAll: jest.fn(),
    retry: jest.fn(),
    retryFailed: jest.fn(),
    clearFinished: jest.fn(),
});

const renderPanel = (queue: UploadQueue) =>
    render(
        <UploadQueueContextProvider value={queue}>
            <UploadQueuePanel />
        </UploadQueueContextProvider>
    );

describe('UploadQueuePanel', () => {
    it('renders without any provider or mock and stays hidden while the queue is empty', () => {
        const { container } = render(<UploadQueuePanel />);

        expect(container).toBeEmptyDOMElement();
    });

    it('shows each file with its status and the finished counter', () => {
        renderPanel(
            buildQueue([
                buildItem({ id: 'u1', displayName: 'Album/a.mp3', status: 'uploading', progress: 40 }),
                buildItem({ id: 'u2', displayName: 'b.txt', status: 'done', progress: 100 }),
                buildItem({ id: 'u3', displayName: 'c.txt', status: 'skipped' }),
                buildItem({ id: 'u4', displayName: 'd.txt', status: 'queued' }),
                buildItem({ id: 'u5', displayName: 'e.txt', status: 'renamed', savedName: 'e (2).txt' }),
            ])
        );

        expect(screen.getByText('FILES_UPLOAD_PANEL_TITLE 3/5')).toBeInTheDocument();
        expect(screen.getByText('Album/a.mp3')).toBeInTheDocument();
        expect(screen.getByRole('progressbar', { name: 'Album/a.mp3' })).toHaveAttribute(
            'aria-valuenow',
            '40'
        );
        expect(screen.getByText('FILES_UPLOAD_STATUS_UPLOADING')).toBeInTheDocument();
        expect(screen.getByText('FILES_UPLOAD_STATUS_DONE')).toBeInTheDocument();
        expect(screen.getByText('FILES_UPLOAD_STATUS_SKIPPED')).toBeInTheDocument();
        expect(screen.getByText('FILES_UPLOAD_STATUS_QUEUED')).toBeInTheDocument();
        expect(screen.getByText('FILES_UPLOAD_STATUS_RENAMED: e (2).txt')).toBeInTheDocument();
    });

    it('renders the backend error verbatim for a failed file and falls back when absent', () => {
        renderPanel(
            buildQueue([
                buildItem({ id: 'u1', displayName: 'a.txt', status: 'failed', error: 'Mensagem do servidor' }),
                buildItem({ id: 'u2', displayName: 'b.txt', status: 'failed' }),
            ])
        );

        expect(screen.getByText('FILES_UPLOAD_STATUS_FAILED: Mensagem do servidor')).toBeInTheDocument();
        expect(screen.getByText('FILES_UPLOAD_STATUS_FAILED: ERROR_UPLOAD_FAILED')).toBeInTheDocument();
    });

    it('wires cancel, cancel all, retry, retry failed and clear finished', () => {
        const queue = buildQueue([
            buildItem({ id: 'u1', displayName: 'a.txt', status: 'uploading' }),
            buildItem({ id: 'u2', displayName: 'b.txt', status: 'failed' }),
            buildItem({ id: 'u3', displayName: 'c.txt', status: 'canceled' }),
        ]);
        renderPanel(queue);

        fireEvent.click(screen.getByRole('button', { name: 'FILES_UPLOAD_CANCEL a.txt' }));
        fireEvent.click(screen.getByRole('button', { name: 'FILES_UPLOAD_RETRY b.txt' }));
        fireEvent.click(screen.getByRole('button', { name: 'FILES_UPLOAD_RETRY c.txt' }));
        fireEvent.click(screen.getByRole('button', { name: 'FILES_UPLOAD_CANCEL_ALL' }));
        fireEvent.click(screen.getByRole('button', { name: 'FILES_UPLOAD_RETRY_FAILED' }));
        fireEvent.click(screen.getByRole('button', { name: 'FILES_UPLOAD_CLEAR_FINISHED' }));

        expect(queue.cancel).toHaveBeenCalledWith('u1');
        expect(queue.retry).toHaveBeenCalledWith('u2');
        expect(queue.retry).toHaveBeenCalledWith('u3');
        expect(queue.cancelAll).toHaveBeenCalled();
        expect(queue.retryFailed).toHaveBeenCalled();
        expect(queue.clearFinished).toHaveBeenCalled();
    });

    it('hides bulk actions that do not apply', () => {
        renderPanel(buildQueue([buildItem({ status: 'queued' })]));

        expect(screen.queryByRole('button', { name: 'FILES_UPLOAD_RETRY_FAILED' })).not.toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'FILES_UPLOAD_CLEAR_FINISHED' })).not.toBeInTheDocument();
    });

    it('collapses and expands the body', () => {
        renderPanel(buildQueue([buildItem({ displayName: 'a.txt' })]));

        fireEvent.click(screen.getByRole('button', { name: 'FILES_UPLOAD_COLLAPSE' }));
        expect(screen.queryByText('a.txt')).not.toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'FILES_UPLOAD_EXPAND' }));
        expect(screen.getByText('a.txt')).toBeInTheDocument();
    });
});
