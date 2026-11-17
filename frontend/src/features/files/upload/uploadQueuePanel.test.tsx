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
    size: 1,
    sentBytes: 0,
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
                buildItem({
                    id: 'u1',
                    displayName: 'Album/a.mp3',
                    status: 'uploading',
                    progress: 40,
                }),
                buildItem({ id: 'u2', displayName: 'b.txt', status: 'done', progress: 100 }),
                buildItem({ id: 'u3', displayName: 'c.txt', status: 'skipped' }),
                buildItem({ id: 'u4', displayName: 'd.txt', status: 'queued' }),
                buildItem({
                    id: 'u5',
                    displayName: 'e.txt',
                    status: 'renamed',
                    savedName: 'e (2).txt',
                }),
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

    it('announces upload progress politely through the panel title', () => {
        renderPanel(buildQueue([buildItem({ id: 'u1', status: 'done', progress: 100 })]));

        const announcement = screen.getByRole('status');
        expect(announcement).toHaveAttribute('aria-live', 'polite');
        expect(announcement).toHaveTextContent('FILES_UPLOAD_PANEL_TITLE 1/1');
    });

    it('renders the backend error verbatim for a failed file and falls back when absent', () => {
        renderPanel(
            buildQueue([
                buildItem({
                    id: 'u1',
                    displayName: 'a.txt',
                    status: 'failed',
                    error: 'Mensagem do servidor',
                }),
                buildItem({ id: 'u2', displayName: 'b.txt', status: 'failed' }),
            ])
        );

        expect(
            screen.getByText('FILES_UPLOAD_STATUS_FAILED: Mensagem do servidor')
        ).toBeInTheDocument();
        expect(
            screen.getByText('FILES_UPLOAD_STATUS_FAILED: ERROR_UPLOAD_FAILED')
        ).toBeInTheDocument();
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

        expect(
            screen.queryByRole('button', { name: 'FILES_UPLOAD_RETRY_FAILED' })
        ).not.toBeInTheDocument();
        expect(
            screen.queryByRole('button', { name: 'FILES_UPLOAD_CLEAR_FINISHED' })
        ).not.toBeInTheDocument();
    });

    it('collapses and expands the body', () => {
        renderPanel(buildQueue([buildItem({ displayName: 'a.txt' })]));

        fireEvent.click(screen.getByRole('button', { name: 'FILES_UPLOAD_COLLAPSE' }));
        expect(screen.queryByText('a.txt')).not.toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'FILES_UPLOAD_EXPAND' }));
        expect(screen.getByText('a.txt')).toBeInTheDocument();
    });

    it('shows the size of each file formatted', () => {
        renderPanel(buildQueue([buildItem({ displayName: 'big.bin', size: 5 * 1024 * 1024 })]));

        expect(screen.getByText('5.00 MB')).toBeInTheDocument();
    });

    it('shows the checking and waiting for connection states, with an indeterminate bar while checking', () => {
        renderPanel(
            buildQueue([
                buildItem({ id: 'u1', displayName: 'a.bin', status: 'checking' }),
                buildItem({
                    id: 'u2',
                    displayName: 'b.bin',
                    status: 'waitingConnection',
                    progress: 30,
                }),
            ])
        );

        expect(screen.getByText('FILES_UPLOAD_STATUS_CHECKING')).toBeInTheDocument();
        expect(screen.getByText('FILES_UPLOAD_STATUS_WAITING_CONNECTION')).toBeInTheDocument();
        expect(screen.getByRole('progressbar', { name: 'a.bin' })).not.toHaveAttribute(
            'aria-valuenow'
        );
        expect(screen.getByRole('progressbar', { name: 'b.bin' })).toHaveAttribute(
            'aria-valuenow',
            '30'
        );
        expect(
            screen.getByRole('button', { name: 'FILES_UPLOAD_CANCEL a.bin' })
        ).toBeInTheDocument();
        expect(
            screen.getByRole('button', { name: 'FILES_UPLOAD_CANCEL b.bin' })
        ).toBeInTheDocument();
    });

    it('shows already existed as a finished state without retry', () => {
        renderPanel(
            buildQueue([
                buildItem({ displayName: 'a.bin', status: 'alreadyExisted', progress: 100 }),
            ])
        );

        expect(screen.getByText('FILES_UPLOAD_STATUS_ALREADY_EXISTED')).toBeInTheDocument();
        expect(screen.getByText('FILES_UPLOAD_PANEL_TITLE 1/1')).toBeInTheDocument();
        expect(
            screen.queryByRole('button', { name: 'FILES_UPLOAD_RETRY a.bin' })
        ).not.toBeInTheDocument();
    });

    it('translates the connection lost key instead of showing backend text', () => {
        renderPanel(
            buildQueue([buildItem({ status: 'failed', errorKey: 'ERROR_UPLOAD_CONNECTION_LOST' })])
        );

        expect(
            screen.getByText('FILES_UPLOAD_STATUS_FAILED: ERROR_UPLOAD_CONNECTION_LOST')
        ).toBeInTheDocument();
    });

    it('shows no speed line while uploading without an estimate or when not uploading', () => {
        renderPanel(
            buildQueue([
                buildItem({ id: 'u1', status: 'uploading' }),
                buildItem({ id: 'u2', status: 'waitingConnection', bytesPerSecond: 1024 }),
            ])
        );

        expect(screen.queryByText(/FILES_UPLOAD_SPEED/)).not.toBeInTheDocument();
    });
});
