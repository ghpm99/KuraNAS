import { render, screen } from '@testing-library/react';
import UploadQueuePanel from './uploadQueuePanel';
import { UploadQueueContextProvider } from './uploadQueueContext';
import type { UploadItem, UploadQueue } from './uploadQueueTypes';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, variables: Record<string, string> = {}) =>
            `${key}(${Object.values(variables).join(',')})`,
    }),
}));

const uploadingItem: UploadItem = {
    id: 'upload-1',
    file: new File(['x'], 'big.bin'),
    displayName: 'big.bin',
    onConflict: 'rename',
    status: 'uploading',
    size: 10 * 1024 * 1024,
    sentBytes: 2 * 1024 * 1024,
    progress: 20,
};

const renderPanelWith = (item: UploadItem) => {
    const queue: UploadQueue = {
        items: [item],
        conflictPolicy: 'rename',
        setConflictPolicy: jest.fn(),
        enqueue: jest.fn(),
        cancel: jest.fn(),
        cancelAll: jest.fn(),
        retry: jest.fn(),
        retryFailed: jest.fn(),
        clearFinished: jest.fn(),
    };
    render(
        <UploadQueueContextProvider value={queue}>
            <UploadQueuePanel />
        </UploadQueueContextProvider>
    );
};

describe('UploadQueuePanel transfer details', () => {
    it('shows the formatted speed and the short remaining time while uploading', () => {
        renderPanelWith({
            ...uploadingItem,
            bytesPerSecond: 2 * 1024 * 1024,
            remainingSeconds: 125,
        });

        expect(
            screen.getByText(
                'FILES_UPLOAD_SPEED(2.00 MB) · FILES_UPLOAD_REMAINING(FILES_UPLOAD_TIME_MINUTES(2))'
            )
        ).toBeInTheDocument();
        expect(screen.getByText('10.00 MB')).toBeInTheDocument();
    });

    it('shows only the speed when the remaining time is unknown', () => {
        renderPanelWith({ ...uploadingItem, bytesPerSecond: 512 * 1024 });

        expect(screen.getByText('FILES_UPLOAD_SPEED(512.00 KB)')).toBeInTheDocument();
    });
});
