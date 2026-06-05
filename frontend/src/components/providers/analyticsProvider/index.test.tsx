import { act, render, screen, waitFor } from '@testing-library/react';
import { AnalyticsProvider } from './index';
import { useAnalyticsOverview } from './analyticsContext';
import { apiBase } from '@/service';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

jest.mock('@/service', () => ({
    apiBase: {
        get: jest.fn(),
    },
}));

const mockedApiGet = apiBase.get as jest.Mock;

const createQueryClient = () =>
    new QueryClient({
        defaultOptions: {
            queries: {
                retry: false,
            },
        },
    });

const routeResponse = (url: string) => {
    switch (url) {
        case '/analytics/storage':
            return {
                data: {
                    storage: { total_bytes: 1000, used_bytes: 500, free_bytes: 500, growth_bytes: 10 },
                    counts: { files_total: 12, files_added: 2, folders: 4 },
                },
            };
        case '/analytics/duplicates':
            return { data: { groups: 0, files: 0, reclaimable_size: 0 } };
        case '/analytics/library':
            return {
                data: {
                    categorized_media: 6,
                    audio_with_metadata: 2,
                    video_with_metadata: 2,
                    image_with_metadata: 2,
                    image_classified: 2,
                },
            };
        case '/analytics/processing':
            return {
                data: {
                    metadata_pending: 1,
                    metadata_failed: 0,
                    thumbnail_pending: 2,
                    thumbnail_failed: 0,
                    recurring_timeouts: 0,
                },
            };
        case '/analytics/health':
            return {
                data: {
                    status: 'ok',
                    last_scan_at: '',
                    last_scan_seconds: 0,
                    indexed_files: 12,
                    errors_last_24h: 0,
                    recent_errors: [],
                },
            };
        case '/analytics/ai-usage':
            return {
                data: {
                    total: 5,
                    success: 4,
                    failure: 1,
                    total_tokens: 100,
                    avg_latency_ms: 120,
                },
            };
        default:
            // timeseries, types, extensions, recent-files, top-folders,
            // hot-folders, duplicates/groups all return arrays.
            return { data: [] };
    }
};

const Consumer = () => {
    const { period, loading, error, data, setPeriod, refresh } = useAnalyticsOverview();
    return (
        <div>
            <span data-testid="period">{period}</span>
            <span data-testid="loading">{loading ? 'yes' : 'no'}</span>
            <span data-testid="error">{error}</span>
            <span data-testid="files">{data?.counts.files_total ?? 0}</span>
            <button onClick={() => setPeriod('30d')}>period</button>
            <button onClick={() => void refresh()}>refresh</button>
        </div>
    );
};

describe('providers/analyticsProvider', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApiGet.mockImplementation((url: string) => Promise.resolve(routeResponse(url)));
    });

    it('loads default period and composes data from slice endpoints', async () => {
        const queryClient = createQueryClient();
        render(
            <QueryClientProvider client={queryClient}>
                <AnalyticsProvider>
                    <Consumer />
                </AnalyticsProvider>
            </QueryClientProvider>
        );

        await waitFor(() => expect(screen.getByTestId('loading')).toHaveTextContent('no'));
        expect(screen.getByTestId('period')).toHaveTextContent('7d');
        expect(screen.getByTestId('files')).toHaveTextContent('12');

        act(() => {
            screen.getByText('period').click();
        });

        await waitFor(() => expect(screen.getByTestId('period')).toHaveTextContent('30d'));
        expect(mockedApiGet).toHaveBeenCalledWith('/analytics/storage', {
            params: { period: '30d' },
        });
    });

    it('exposes error key when the storage slice fails', async () => {
        mockedApiGet.mockImplementation((url: string) => {
            if (url === '/analytics/storage') {
                return Promise.reject(new Error('network'));
            }
            return Promise.resolve(routeResponse(url));
        });
        const queryClient = createQueryClient();
        render(
            <QueryClientProvider client={queryClient}>
                <AnalyticsProvider>
                    <Consumer />
                </AnalyticsProvider>
            </QueryClientProvider>
        );

        await waitFor(() => expect(screen.getByTestId('loading')).toHaveTextContent('no'));
        expect(screen.getByTestId('error')).toHaveTextContent('ANALYTICS_ERROR_LOAD_BLOCK');
    });
});
