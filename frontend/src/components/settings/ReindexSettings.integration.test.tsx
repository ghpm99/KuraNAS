import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import { apiBase } from '@/service';
import IndexingSettingsSection from './IndexingSettingsSection';
import StorageRootsSettingsSection from './StorageRootsSettingsSection';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock };

const indexing = {
    scan_on_startup: true,
    extract_metadata: true,
    generate_previews: true,
    index_document_text: true,
    scan_interval_hours: 24,
};

const wrap = (children: React.ReactNode) => (
    <QueryClientProvider
        client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
        <SnackbarProvider>{children}</SnackbarProvider>
    </QueryClientProvider>
);

const jobWithStatus = (status: string) => ({
    data: { id: 7, status, progress: { progress: 0 }, last_error: '' },
});

describe('settings reindex (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.post.mockResolvedValue({ data: { job_id: 7 } });
        mockedApi.get.mockImplementation((url: string) => {
            if (url === '/storage-roots') {
                return Promise.resolve({
                    data: [
                        { id: 1, path: '/a', label: 'A', enabled: true, created_at: '' },
                        { id: 2, path: '/b', label: 'B', enabled: true, created_at: '' },
                    ],
                });
            }
            if (url === '/jobs/7') return Promise.resolve(jobWithStatus('completed'));
            return Promise.reject(new Error(`unexpected GET ${url}`));
        });
    });

    it('confirmed "reindex everything" posts scope all and re-enables the button when the job completes', async () => {
        let resolveJob: (value: unknown) => void = () => undefined;
        mockedApi.get.mockImplementation(
            (url: string) =>
                new Promise((resolve) => {
                    resolveJob = () => resolve(jobWithStatus('completed'));
                    if (url !== '/jobs/7') resolve({ data: [] });
                })
        );
        render(
            wrap(
                <IndexingSettingsSection
                    indexing={indexing}
                    isWorkersEnabled
                    isDisabled={false}
                    onChange={jest.fn()}
                />
            )
        );

        fireEvent.click(screen.getByRole('button', { name: 'SETTINGS_REINDEX_ALL' }));
        expect(mockedApi.post).not.toHaveBeenCalled();
        fireEvent.click(screen.getByText('SETTINGS_REINDEX_CONFIRM'));

        await waitFor(() =>
            expect(mockedApi.post).toHaveBeenCalledWith('/indexing/reindex', { scope: 'all' })
        );
        expect(
            await screen.findByRole('button', { name: 'SETTINGS_REINDEX_RUNNING' })
        ).toBeDisabled();

        await waitFor(() => expect(mockedApi.get).toHaveBeenCalledWith('/jobs/7'));
        resolveJob(null);
        expect(await screen.findByRole('button', { name: 'SETTINGS_REINDEX_ALL' })).toBeEnabled();
    });

    it('confirmed "reindex this root" posts scope root with the root id', async () => {
        render(wrap(<StorageRootsSettingsSection />));

        const rootButtons = await screen.findAllByRole('button', { name: 'SETTINGS_REINDEX_ROOT' });
        fireEvent.click(rootButtons[1] as HTMLElement);
        fireEvent.click(screen.getByText('SETTINGS_REINDEX_CONFIRM'));

        await waitFor(() =>
            expect(mockedApi.post).toHaveBeenCalledWith('/indexing/reindex', {
                scope: 'root',
                root_id: 2,
            })
        );
    });
});
