import { fireEvent, render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import UploadConflictDialog from './uploadConflictDialog';
import UploadQueueProvider from './uploadQueueProvider';
import { UploadQueueContextProvider } from './uploadQueueContext';
import type { UploadItem, UploadQueue } from './uploadQueueTypes';

const buildItem = (overrides: Partial<UploadItem>): UploadItem => ({
    id: 'upload-1',
    file: new File(['x'], 'a.txt'),
    displayName: 'a.txt',
    onConflict: 'ask',
    status: 'awaitingDecision',
    size: 1,
    sentBytes: 0,
    progress: 0,
    ...overrides,
});

const buildQueue = (items: UploadItem[]): UploadQueue => ({
    items,
    conflictPolicy: 'ask',
    setConflictPolicy: jest.fn(),
    enqueue: jest.fn(),
    cancel: jest.fn(),
    cancelAll: jest.fn(),
    retry: jest.fn(),
    retryFailed: jest.fn(),
    clearFinished: jest.fn(),
    resolveConflict: jest.fn(),
    dismissConflict: jest.fn(),
});

const renderDialog = (queue: UploadQueue) =>
    render(
        <UploadQueueContextProvider value={queue}>
            <UploadConflictDialog />
        </UploadQueueContextProvider>
    );

describe('UploadConflictDialog', () => {
    it('renders without any provider and stays closed', () => {
        render(<UploadConflictDialog />);

        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('stays closed inside the real provider while the queue is empty', () => {
        render(
            <QueryClientProvider client={new QueryClient()}>
                <UploadQueueProvider>
                    <span>child</span>
                </UploadQueueProvider>
            </QueryClientProvider>
        );

        expect(screen.getByText('child')).toBeInTheDocument();
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('stays closed when no item is awaiting a decision', () => {
        renderDialog(buildQueue([buildItem({ status: 'uploading' })]));

        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('opens for the first item awaiting a decision', () => {
        renderDialog(
            buildQueue([
                buildItem({ id: 'u1', status: 'done' }),
                buildItem({ id: 'u2', displayName: 'b.txt' }),
                buildItem({ id: 'u3', displayName: 'c.txt' }),
            ])
        );

        expect(screen.getByRole('dialog')).toBeInTheDocument();
        expect(screen.getByText('FILES_UPLOAD_CONFLICT_TITLE')).toBeInTheDocument();
        expect(screen.getByText('FILES_UPLOAD_CONFLICT_MESSAGE')).toBeInTheDocument();
    });

    it.each([
        ['FILES_UPLOAD_ON_CONFLICT_RENAME', 'rename'],
        ['FILES_UPLOAD_ON_CONFLICT_REPLACE', 'replace'],
        ['FILES_UPLOAD_ON_CONFLICT_SKIP', 'skip'],
    ])('the %s button resolves the conflict with %s', (buttonLabel, choice) => {
        const queue = buildQueue([buildItem({ id: 'u1' })]);
        renderDialog(queue);

        fireEvent.click(screen.getByRole('button', { name: buttonLabel }));

        expect(queue.resolveConflict).toHaveBeenCalledWith('u1', choice, false);
        expect(queue.dismissConflict).not.toHaveBeenCalled();
    });

    it('shows apply to all only with more than one item to cover', () => {
        renderDialog(buildQueue([buildItem({ id: 'u1' })]));

        expect(screen.queryByRole('checkbox')).not.toBeInTheDocument();
    });

    it('counts queued items still asking and ignores queued items with a chosen policy', () => {
        const withAskingQueued = buildQueue([
            buildItem({ id: 'u1' }),
            buildItem({ id: 'u2', status: 'queued', onConflict: 'ask' }),
            buildItem({ id: 'u3', status: 'queued', onConflict: 'skip' }),
        ]);
        renderDialog(withAskingQueued);

        expect(screen.getByRole('checkbox')).toBeInTheDocument();
        expect(screen.getByText('FILES_UPLOAD_CONFLICT_APPLY_TO_ALL')).toBeInTheDocument();
    });

    it('does not offer apply to all when the other queued items already have a policy', () => {
        renderDialog(
            buildQueue([
                buildItem({ id: 'u1' }),
                buildItem({ id: 'u2', status: 'queued', onConflict: 'skip' }),
            ])
        );

        expect(screen.queryByRole('checkbox')).not.toBeInTheDocument();
    });

    it('passes apply to all when checked', () => {
        const queue = buildQueue([buildItem({ id: 'u1' }), buildItem({ id: 'u2' })]);
        renderDialog(queue);

        fireEvent.click(screen.getByRole('checkbox'));
        fireEvent.click(screen.getByRole('button', { name: 'FILES_UPLOAD_ON_CONFLICT_REPLACE' }));

        expect(queue.resolveConflict).toHaveBeenCalledWith('u1', 'replace', true);
    });

    it('dismisses the upload with the cancel button', () => {
        const queue = buildQueue([buildItem({ id: 'u1' })]);
        renderDialog(queue);

        fireEvent.click(screen.getByRole('button', { name: 'FILES_UPLOAD_CONFLICT_CANCEL' }));

        expect(queue.dismissConflict).toHaveBeenCalledWith('u1');
        expect(queue.resolveConflict).not.toHaveBeenCalled();
    });

    it('dismisses the upload when closed with Escape', () => {
        const queue = buildQueue([buildItem({ id: 'u1' })]);
        renderDialog(queue);

        fireEvent.keyDown(screen.getByRole('dialog'), { key: 'Escape' });

        expect(queue.dismissConflict).toHaveBeenCalledWith('u1');
    });

    it('opens for the next item awaiting a decision after the first is resolved', () => {
        const firstQueue = buildQueue([
            buildItem({ id: 'u1', displayName: 'a.txt' }),
            buildItem({ id: 'u2', displayName: 'b.txt' }),
        ]);
        const { rerender } = renderDialog(firstQueue);
        fireEvent.click(screen.getByRole('button', { name: 'FILES_UPLOAD_ON_CONFLICT_SKIP' }));
        expect(firstQueue.resolveConflict).toHaveBeenLastCalledWith('u1', 'skip', false);

        const afterResolving = buildQueue([
            buildItem({ id: 'u1', displayName: 'a.txt', status: 'queued', onConflict: 'skip' }),
            buildItem({ id: 'u2', displayName: 'b.txt' }),
        ]);
        rerender(
            <UploadQueueContextProvider value={afterResolving}>
                <UploadConflictDialog />
            </UploadQueueContextProvider>
        );
        fireEvent.click(screen.getByRole('button', { name: 'FILES_UPLOAD_ON_CONFLICT_REPLACE' }));

        expect(afterResolving.resolveConflict).toHaveBeenCalledWith('u2', 'replace', false);
    });
});
