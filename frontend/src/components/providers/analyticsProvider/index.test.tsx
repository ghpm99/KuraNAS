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

const Consumer = () => {
	const { period, loading, error, data, setPeriod, refresh } = useAnalyticsOverview();
	return (
		<div>
			<span data-testid='period'>{period}</span>
			<span data-testid='loading'>{loading ? 'yes' : 'no'}</span>
			<span data-testid='error'>{error}</span>
			<span data-testid='files'>{data?.counts.files_total ?? 0}</span>
			<button onClick={() => setPeriod('30d')}>period</button>
			<button onClick={() => void refresh()}>refresh</button>
		</div>
	);
};

describe('providers/analyticsProvider', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApiGet.mockResolvedValue({
			data: {
				period: '7d',
				generated_at: new Date().toISOString(),
				storage: { total_bytes: 1000, used_bytes: 500, free_bytes: 500, growth_bytes: 10 },
				counts: { files_total: 12, files_added: 2, folders: 4 },
				time_series: [],
				types: [],
				extensions: [],
				hot_folders: [],
				top_folders: [],
				recent_files: [],
				duplicates: { groups: 0, files: 0, reclaimable_size: 0, top_groups: [] },
				health: {
					status: 'ok',
					last_scan_at: '',
					last_scan_seconds: 0,
					indexed_files: 12,
					errors_last_24h: 0,
					recent_errors: [],
				},
			},
		});
	});

	it('loads default period and updates when period changes', async () => {
		const queryClient = createQueryClient();
		render(
			<QueryClientProvider client={queryClient}>
				<AnalyticsProvider>
					<Consumer />
				</AnalyticsProvider>
			</QueryClientProvider>,
		);

		await waitFor(() => expect(screen.getByTestId('loading')).toHaveTextContent('no'));
		expect(screen.getByTestId('period')).toHaveTextContent('7d');
		expect(screen.getByTestId('files')).toHaveTextContent('12');

		act(() => {
			screen.getByText('period').click();
		});

		await waitFor(() => expect(screen.getByTestId('period')).toHaveTextContent('30d'));
		expect(mockedApiGet).toHaveBeenCalledWith('/analytics/overview', { params: { period: '30d' } });
	});

	it('exposes error key when request fails', async () => {
		mockedApiGet.mockRejectedValueOnce(new Error('network'));
		const queryClient = createQueryClient();
		render(
			<QueryClientProvider client={queryClient}>
				<AnalyticsProvider>
					<Consumer />
				</AnalyticsProvider>
			</QueryClientProvider>,
		);

		await waitFor(() => expect(screen.getByTestId('loading')).toHaveTextContent('no'));
		expect(screen.getByTestId('error')).toHaveTextContent('ANALYTICS_ERROR_LOAD_BLOCK');
	});
});
