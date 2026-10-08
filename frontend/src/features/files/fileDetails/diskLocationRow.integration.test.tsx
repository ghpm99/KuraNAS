import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import DiskLocationRow from './diskLocationRow';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock };

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
});
