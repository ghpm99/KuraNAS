import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import DiskLocationRow from './diskLocationRow';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock };

describe('DiskLocationRow (seam)', () => {
    it('GETs /files/location/:id and renders the disk_path field the backend sends', async () => {
        mockedApi.get.mockResolvedValue({
            data: {
                file_id: 5,
                tier: 'cold',
                logical_path: '/docs/a.txt',
                disk_path: '/cold/main/docs/a.txt',
                logical_disk_path: '/data/docs/a.txt',
                root_label: 'main',
                exists_on_disk: false,
            },
        });

        render(
            <QueryClientProvider client={new QueryClient()}>
                <DiskLocationRow fileId={5} />
            </QueryClientProvider>
        );

        expect(await screen.findByText('/cold/main/docs/a.txt')).toBeInTheDocument();
        expect(screen.getByText('FILE_DISK_LOCATION_MISSING')).toBeInTheDocument();
        expect(mockedApi.get).toHaveBeenCalledWith('/files/location/5');
    });

    it('POSTs /tiering/promote/:id from the cold-file button and refetches the location', async () => {
        const coldLocation = {
            file_id: 5,
            tier: 'cold',
            logical_path: '/docs/a.txt',
            disk_path: '/cold/main/docs/a.txt',
            logical_disk_path: '/data/docs/a.txt',
            root_label: 'main',
            exists_on_disk: true,
        };
        mockedApi.get.mockResolvedValueOnce({ data: coldLocation });
        mockedApi.get.mockResolvedValue({
            data: { ...coldLocation, tier: 'hot', disk_path: '/data/docs/a.txt' },
        });
        mockedApi.post.mockResolvedValue({
            data: { file_id: 5, tier: 'hot', disk_path: '/data/docs/a.txt' },
        });

        render(
            <QueryClientProvider client={new QueryClient()}>
                <DiskLocationRow fileId={5} />
            </QueryClientProvider>
        );

        fireEvent.click(await screen.findByRole('button', { name: 'FILE_PROMOTE_TO_HOT' }));

        await waitFor(() => expect(mockedApi.post).toHaveBeenCalledWith('/tiering/promote/5'));
        await waitFor(() => expect(screen.queryByRole('button', { name: 'FILE_PROMOTE_TO_HOT' })).toBeNull());
        expect(screen.getByText('/data/docs/a.txt')).toBeInTheDocument();
    });
});
