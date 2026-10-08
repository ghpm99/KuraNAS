import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import type { ReactNode } from 'react';
import DiskLocationRow from './diskLocationRow';
import { getFileLocation } from '@/service/files';
import type { FileLocation } from '@/types/fileLocation';

const withProviders = (children: ReactNode) => (
    <QueryClientProvider
        client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
        <SnackbarProvider>{children}</SnackbarProvider>
    </QueryClientProvider>
);

jest.mock('@/service/files', () => ({
    ...jest.requireActual('@/service/files'),
    getFileLocation: jest.fn(),
}));

const mockedGetFileLocation = getFileLocation as jest.Mock;

const coldLocation: FileLocation = {
    file_id: 5,
    tier: 'cold',
    logical_path: '/docs/a.txt',
    disk_path: 'F:\\kuranas-cold\\main\\docs\\a.txt',
    logical_disk_path: 'D:\\data\\docs\\a.txt',
    root_label: 'main',
    exists_on_disk: true,
};

describe('DiskLocationRow with a location', () => {
    const writeText = jest.fn();

    beforeEach(() => {
        jest.clearAllMocks();
        Object.assign(navigator, { clipboard: { writeText } });
    });

    it('shows the disk path and copies it to the clipboard', async () => {
        mockedGetFileLocation.mockResolvedValue(coldLocation);
        writeText.mockResolvedValue(undefined);
        render(withProviders(<DiskLocationRow fileId={5} />));

        expect(await screen.findByText(coldLocation.disk_path)).toBeInTheDocument();
        expect(screen.queryByText('FILE_DISK_LOCATION_MISSING')).toBeNull();

        fireEvent.click(screen.getByRole('button', { name: 'FILE_DISK_LOCATION_COPY' }));

        await waitFor(() => expect(writeText).toHaveBeenCalledWith(coldLocation.disk_path));
        expect(await screen.findByText('FILE_DISK_LOCATION_COPIED')).toBeInTheDocument();
    });

    it('warns when the file is missing from the disk', async () => {
        mockedGetFileLocation.mockResolvedValue({ ...coldLocation, exists_on_disk: false });
        render(withProviders(<DiskLocationRow fileId={5} />));

        expect(await screen.findByText('FILE_DISK_LOCATION_MISSING')).toBeInTheDocument();
    });

    it('reports a copy failure when the clipboard rejects', async () => {
        mockedGetFileLocation.mockResolvedValue(coldLocation);
        writeText.mockRejectedValue(new Error('denied'));
        render(withProviders(<DiskLocationRow fileId={5} />));

        fireEvent.click(await screen.findByRole('button', { name: 'FILE_DISK_LOCATION_COPY' }));

        expect(await screen.findByText('FILE_DISK_LOCATION_COPY_FAILED')).toBeInTheDocument();
    });

    it('reports a copy failure when the clipboard is unavailable', async () => {
        mockedGetFileLocation.mockResolvedValue(coldLocation);
        Object.assign(navigator, { clipboard: undefined });
        render(withProviders(<DiskLocationRow fileId={5} />));

        fireEvent.click(await screen.findByRole('button', { name: 'FILE_DISK_LOCATION_COPY' }));

        expect(await screen.findByText('FILE_DISK_LOCATION_COPY_FAILED')).toBeInTheDocument();
    });

    it('hides the row when the location request fails', async () => {
        mockedGetFileLocation.mockRejectedValue(new Error('boom'));
        const { container } = render(withProviders(<DiskLocationRow fileId={5} />));

        await waitFor(() => expect(mockedGetFileLocation).toHaveBeenCalled());
        expect(container.querySelector('code')).toBeNull();
    });
});
