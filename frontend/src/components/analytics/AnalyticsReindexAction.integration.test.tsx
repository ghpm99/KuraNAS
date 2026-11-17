import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import { MemoryRouter } from 'react-router-dom';
import AnalyticsReindexAction from './AnalyticsReindexAction';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: {
        get: jest.fn(),
        post: jest.fn(),
    },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock };

const renderAction = (onFinished: () => void) => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <SnackbarProvider>
                <MemoryRouter>
                    <AnalyticsReindexAction onFinished={onFinished} />
                </MemoryRouter>
            </SnackbarProvider>
        </QueryClientProvider>
    );
};

describe('components/analytics/AnalyticsReindexAction (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.post.mockResolvedValue({ data: { job_id: 7 } });
        mockedApi.get.mockImplementation((url: string) => {
            if (url === '/jobs/7') {
                return Promise.resolve({ data: { id: 7, status: 'completed', progress: {} } });
            }
            return Promise.reject(new Error(`unexpected GET ${url}`));
        });
    });

    it('clicking Reindex posts scope all, and reports completion', async () => {
        const onFinished = jest.fn();
        renderAction(onFinished);

        fireEvent.click(screen.getByRole('button', { name: 'ANALYTICS_ACTION_REINDEX' }));

        await waitFor(() => expect(mockedApi.post).toHaveBeenCalledTimes(1));
        expect(mockedApi.post).toHaveBeenCalledWith('/indexing/reindex', { scope: 'all' });
        await waitFor(() => expect(onFinished).toHaveBeenCalledTimes(1));
    });
});
