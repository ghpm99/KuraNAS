import { act, render, screen } from '@testing-library/react';
import { AnalyticsProvider, useAnalytics } from './AnalyticsContext';
import { useQuery } from '@tanstack/react-query';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
	apiBase: {
		get: jest.fn(),
	},
}));

jest.mock('@tanstack/react-query', () => ({
	useQuery: jest.fn(),
}));

const mockedUseQuery = useQuery as jest.Mock;
const mockedApiGet = apiBase.get as jest.Mock;

function Consumer() {
	const { analyticsData, refreshAnalytics } = useAnalytics();
	return (
		<div>
			<span data-testid='space'>{analyticsData.storageOverview.totalUsedSpace}</span>
			<span data-testid='files'>{analyticsData.storageOverview.totalFiles}</span>
			<span data-testid='folders'>{analyticsData.storageOverview.totalFolders}</span>
			<span data-testid='types'>{analyticsData.fileTypes.length}</span>
			<span data-testid='largest'>{analyticsData.largestFiles.length}</span>
			<span data-testid='dups'>{analyticsData.duplicates.total ?? 0}</span>
			<button onClick={refreshAnalytics}>refresh</button>
		</div>
	);
}

describe('contexts/AnalyticsContext', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApiGet.mockResolvedValue({ data: { total_space_used: 1024 } });
		mockedUseQuery
			.mockReturnValueOnce({ data: '1.00 KB', refetch: jest.fn() })
			.mockReturnValueOnce({ data: 10, refetch: jest.fn() })
			.mockReturnValueOnce({ data: 3, refetch: jest.fn() })
			.mockReturnValueOnce({ data: [{ format: '.mp3', total: 2, size: 200, percentage: 10 }], refetch: jest.fn() })
			.mockReturnValueOnce({ data: [{ id: 1, name: 'a', size: 10, path: '/a' }], refetch: jest.fn() })
			.mockReturnValueOnce({ data: { total: 0, total_size: 0, files: [] }, refetch: jest.fn() });
	});

	it('throws when hook is used outside provider', () => {
		expect(() => render(<Consumer />)).toThrow('useAnalytics must be used within an AnalyticsProvider');
	});

	it('provides analytics data and refresh function', () => {
		render(
			<AnalyticsProvider>
				<Consumer />
			</AnalyticsProvider>,
		);

		expect(screen.getByTestId('space')).toHaveTextContent('1.00 KB');
		expect(screen.getByTestId('files')).toHaveTextContent('10');

		const refetchFns = mockedUseQuery.mock.results.map((r) => r.value.refetch);
		act(() => {
			screen.getByText('refresh').click();
		});
		refetchFns.forEach((fn: jest.Mock) => expect(fn).toHaveBeenCalled());
	});

	it('runs query functions including empty totalUsedSpace branch', async () => {
		render(
			<AnalyticsProvider>
				<Consumer />
			</AnalyticsProvider>,
		);

		const opts = mockedUseQuery.mock.calls.map((c) => c[0]);
		mockedApiGet.mockResolvedValueOnce({ data: {} });
		await expect(opts[0].queryFn()).resolves.toBe('');
		mockedApiGet.mockResolvedValueOnce({ data: null });
		await expect(opts[0].queryFn()).resolves.toBe('');

		mockedApiGet
			.mockResolvedValueOnce({ data: { total_files: 9 } })
			.mockResolvedValueOnce({ data: { total_directory: 4 } })
			.mockResolvedValueOnce({ data: [{ format: '.png', total: 1, size: 100, percentage: 5 }] })
			.mockResolvedValueOnce({ data: [{ id: 2 }] })
			.mockResolvedValueOnce({ data: { total: 1, total_size: 100, files: [] } });

		await expect(opts[1].queryFn()).resolves.toBe(9);
		await expect(opts[2].queryFn()).resolves.toBe(4);
		await expect(opts[3].queryFn()).resolves.toEqual([{ format: '.png', total: 1, size: 100, percentage: 5 }]);
		await expect(opts[4].queryFn()).resolves.toEqual([{ id: 2 }]);
		await expect(opts[5].queryFn()).resolves.toEqual({ total: 1, total_size: 100, files: [] });
	});

	it('throws directly when useAnalytics hook executes without provider', () => {
		function Bare() {
			useAnalytics();
			return null;
		}
		expect(() => render(<Bare />)).toThrow('useAnalytics must be used within an AnalyticsProvider');
	});

	it('uses fallback values when queries return undefined', () => {
		mockedUseQuery.mockReset();
		mockedUseQuery
			.mockReturnValueOnce({ data: undefined, refetch: jest.fn() })
			.mockReturnValueOnce({ data: undefined, refetch: jest.fn() })
			.mockReturnValueOnce({ data: undefined, refetch: jest.fn() })
			.mockReturnValueOnce({ data: undefined, refetch: jest.fn() })
			.mockReturnValueOnce({ data: undefined, refetch: jest.fn() })
			.mockReturnValueOnce({ data: undefined, refetch: jest.fn() });

		render(
			<AnalyticsProvider>
				<Consumer />
			</AnalyticsProvider>,
		);

		expect(screen.getByTestId('space')).toHaveTextContent('');
		expect(screen.getByTestId('files')).toHaveTextContent('0');
		expect(screen.getByTestId('folders')).toHaveTextContent('0');
		expect(screen.getByTestId('types')).toHaveTextContent('0');
		expect(screen.getByTestId('largest')).toHaveTextContent('0');
		expect(screen.getByTestId('dups')).toHaveTextContent('0');
	});
});
