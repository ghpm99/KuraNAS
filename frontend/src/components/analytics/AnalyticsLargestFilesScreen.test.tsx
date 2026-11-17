import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import AnalyticsLargestFilesScreen from './AnalyticsLargestFilesScreen';
import { getLargestFilesPage } from '@/service/files';
import { createTestFile } from '@/features/files/selection/testFileFactory';
import { buildServerPage } from '@/shared/test/buildServerPage';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';

jest.mock('@/service/files', () => ({ getLargestFilesPage: jest.fn() }));

const mockedGetLargestFilesPage = getLargestFilesPage as jest.Mock;

const bigVideo = createTestFile(1, {
    name: 'filme.mkv',
    path: '/Videos/filme.mkv',
    parent_path: '/Videos',
    size: 2 * 1024 * 1024 * 1024,
    format: '.mkv',
});
const smallPhoto = createTestFile(2, {
    name: 'foto.jpg',
    path: '/Fotos/foto.jpg',
    parent_path: '/Fotos',
    size: 2048,
    format: '.jpg',
});

describe('AnalyticsLargestFilesScreen', () => {
    beforeEach(() => jest.clearAllMocks());

    it('lists name, folder, size and type and links to the file and its folder', async () => {
        mockedGetLargestFilesPage.mockResolvedValue(
            buildServerPage([bigVideo, smallPhoto], { page: 1, pageSize: 50, hasNext: false })
        );

        renderWithoutBackend(<AnalyticsLargestFilesScreen />);

        const videoRow = (await screen.findByText('filme.mkv')).closest('tr') as HTMLElement;
        expect(within(videoRow).getByText('2.00 GB')).toBeInTheDocument();
        expect(within(videoRow).getByText('.mkv')).toBeInTheDocument();
        expect(within(videoRow).getByRole('link', { name: 'filme.mkv' })).toHaveAttribute(
            'href',
            '/files/Videos/filme.mkv'
        );
        expect(within(videoRow).getByRole('link', { name: '/Videos' })).toHaveAttribute(
            'href',
            '/files/Videos'
        );
        expect(mockedGetLargestFilesPage).toHaveBeenCalledWith({ page: 1, pageSize: 50 });
    });

    it('shows the empty state when there are no files', async () => {
        mockedGetLargestFilesPage.mockResolvedValue(
            buildServerPage([], { page: 1, pageSize: 50, hasNext: false })
        );

        renderWithoutBackend(<AnalyticsLargestFilesScreen />);

        expect(await screen.findByText('ANALYTICS_LARGEST_EMPTY')).toBeInTheDocument();
    });

    it('requests the next page while has_next is true and stops when it is false', async () => {
        mockedGetLargestFilesPage
            .mockResolvedValueOnce(
                buildServerPage([bigVideo], { page: 1, pageSize: 50, hasNext: true })
            )
            .mockResolvedValueOnce(
                buildServerPage([smallPhoto], { page: 2, pageSize: 50, hasNext: false })
            );

        renderWithoutBackend(<AnalyticsLargestFilesScreen />);
        fireEvent.click(await screen.findByRole('button', { name: 'LOAD_MORE' }));

        expect(await screen.findByText('foto.jpg')).toBeInTheDocument();
        expect(mockedGetLargestFilesPage).toHaveBeenLastCalledWith({ page: 2, pageSize: 50 });
        expect(screen.getByText('filme.mkv')).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'LOAD_MORE' })).not.toBeInTheDocument();
    });

    it('keeps the loaded rows when the next page fails and retries that page', async () => {
        mockedGetLargestFilesPage
            .mockResolvedValueOnce(
                buildServerPage([bigVideo], { page: 1, pageSize: 50, hasNext: true })
            )
            .mockRejectedValueOnce({ response: { data: { error: 'falha do servidor' } } })
            .mockResolvedValueOnce(
                buildServerPage([smallPhoto], { page: 2, pageSize: 50, hasNext: false })
            );

        renderWithoutBackend(<AnalyticsLargestFilesScreen />);
        fireEvent.click(await screen.findByRole('button', { name: 'LOAD_MORE' }));

        expect(await screen.findByText('falha do servidor')).toBeInTheDocument();
        expect(screen.getByText('filme.mkv')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: /TRY_AGAIN|Tentar novamente/ }));

        expect(await screen.findByText('foto.jpg')).toBeInTheDocument();
        await waitFor(() =>
            expect(mockedGetLargestFilesPage).toHaveBeenLastCalledWith({ page: 2, pageSize: 50 })
        );
        expect(screen.getByText('filme.mkv')).toBeInTheDocument();
    });
});
