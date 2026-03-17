import { fireEvent, render, screen } from '@testing-library/react';
import AnalyticsOverviewScreen from './AnalyticsOverviewScreen';
import type { AnalyticsScreenState } from './useAnalyticsScreenState';
import type { AnalyticsOverview } from '@/types/analytics';

const mockNavigate = jest.fn();

jest.mock('react-router-dom', () => ({
	useNavigate: () => mockNavigate,
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const createOverview = (overrides?: Partial<AnalyticsOverview>): AnalyticsOverview => ({
	period: '7d',
	generated_at: '2026-03-16T11:50:00Z',
	storage: {
		total_bytes: 1000,
		used_bytes: 400,
		free_bytes: 600,
		growth_bytes: 50,
	},
	counts: {
		files_total: 200,
		files_added: 15,
		folders: 10,
	},
	time_series: [
		{ date: '2026-03-15', used_bytes: 380 },
		{ date: '2026-03-16', used_bytes: 400 },
	],
	types: [
		{ type: 'audio', count: 100, bytes: 200 },
		{ type: 'video', count: 50, bytes: 300 },
	],
	extensions: [
		{ ext: '.mp3', count: 80, bytes: 160 },
		{ ext: '.mp4', count: 50, bytes: 300 },
	],
	hot_folders: [{ path: '/media/music', new_files: 5, added_bytes: 100, last_event: '2026-03-16' }],
	top_folders: [{ path: '/media', files: 150, bytes: 500, last_modified: '2026-03-16T10:00:00Z' }],
	recent_files: [],
	duplicates: {
		groups: 3,
		files: 6,
		reclaimable_size: 100,
		top_groups: [{ signature: 'abc123def456', copies: 3, size_bytes: 50, reclaimable_size: 100, paths: ['/a', '/b'] }],
	},
	library: {
		categorized_media: 10,
		audio_with_metadata: 8,
		video_with_metadata: 5,
		image_with_metadata: 3,
		image_classified: 2,
	},
	processing: {
		metadata_pending: 0,
		metadata_failed: 1,
		thumbnail_pending: 0,
		thumbnail_failed: 2,
	},
	health: {
		status: 'ok',
		last_scan_at: '2026-03-16T11:40:00Z',
		last_scan_seconds: 12,
		indexed_files: 200,
		errors_last_24h: 0,
		recent_errors: [],
	},
	...overrides,
});

const createState = (overrides: Partial<AnalyticsScreenState> = {}): AnalyticsScreenState =>
	({
		t: (key: string, params?: Record<string, string | number>) => {
			if (params) {
				let result = key;
				for (const [k, v] of Object.entries(params)) {
					result += ` ${k}=${v}`;
				}
				return result;
			}
			return key;
		},
		period: '7d',
		setPeriod: jest.fn(),
		data: createOverview(),
		loading: false,
		error: '',
		refresh: jest.fn().mockResolvedValue(undefined),
		formatBytes: (n: number) => `${n} B`,
		formatPercent: (n: number) => `${n}%`,
		formatDate: (s: string) => s || '-',
		usedPercent: 40,
		reclaimablePercent: 10,
		healthStatusLabel: 'Healthy',
		healthStatusColor: 'success',
		processingFailureTotal: 3,
		updatedMinutes: '10',
		...overrides,
	}) as unknown as AnalyticsScreenState;

describe('AnalyticsOverviewScreen', () => {
	beforeEach(() => {
		mockNavigate.mockReset();
	});

	it('renders all 6 KPI cards', () => {
		const state = createState();
		render(<AnalyticsOverviewScreen state={state} />);

		expect(screen.getByText('ANALYTICS_KPI_STORAGE')).toBeInTheDocument();
		expect(screen.getByText('ANALYTICS_KPI_GROWTH')).toBeInTheDocument();
		expect(screen.getByText('ANALYTICS_KPI_FILES_ADDED')).toBeInTheDocument();
		expect(screen.getByText('ANALYTICS_KPI_HOT_FOLDERS')).toBeInTheDocument();
		expect(screen.getByText('ANALYTICS_KPI_DUPLICATES')).toBeInTheDocument();
		expect(screen.getByText('ANALYTICS_KPI_INDEX_STATUS')).toBeInTheDocument();
	});

	it('renders storage KPI with formatted bytes', () => {
		const state = createState();
		render(<AnalyticsOverviewScreen state={state} />);

		expect(screen.getByText('400 B / 1000 B')).toBeInTheDocument();
	});

	it('renders growth KPI with positive prefix', () => {
		const state = createState();
		render(<AnalyticsOverviewScreen state={state} />);

		expect(screen.getByText('+50 B')).toBeInTheDocument();
	});

	it('renders files added count', () => {
		const state = createState();
		render(<AnalyticsOverviewScreen state={state} />);

		expect(screen.getByText('15')).toBeInTheDocument();
	});

	it('renders analytics sections', () => {
		const state = createState();
		render(<AnalyticsOverviewScreen state={state} />);

		expect(screen.getByText('ANALYTICS_STORAGE_TREND')).toBeInTheDocument();
		expect(screen.getByText('ANALYTICS_STORAGE_BY_CATEGORY')).toBeInTheDocument();
		expect(screen.getByText('ANALYTICS_TOP_FOLDERS')).toBeInTheDocument();
		expect(screen.getByText('ANALYTICS_EXTENSIONS')).toBeInTheDocument();
		expect(screen.getByText('ANALYTICS_DUPLICATE_FILES')).toBeInTheDocument();
		expect(screen.getByText('ANALYTICS_INDEX_HEALTH')).toBeInTheDocument();
	});

	it('renders time series table rows', () => {
		const state = createState();
		render(<AnalyticsOverviewScreen state={state} />);

		expect(screen.getByText('2026-03-15')).toBeInTheDocument();
		expect(screen.getByText('2026-03-16')).toBeInTheDocument();
	});

	it('renders types table rows', () => {
		const state = createState();
		render(<AnalyticsOverviewScreen state={state} />);

		expect(screen.getByText('audio')).toBeInTheDocument();
		expect(screen.getByText('video')).toBeInTheDocument();
	});

	it('renders top folders with click navigation', () => {
		const state = createState();
		render(<AnalyticsOverviewScreen state={state} />);

		const row = screen.getByText('/media').closest('tr');
		expect(row).toBeInTheDocument();
		fireEvent.click(row!);

		expect(mockNavigate).toHaveBeenCalled();
	});

	it('renders extensions table', () => {
		const state = createState();
		render(<AnalyticsOverviewScreen state={state} />);

		expect(screen.getByText('.mp3')).toBeInTheDocument();
		expect(screen.getByText('.mp4')).toBeInTheDocument();
	});

	it('renders duplicate groups table', () => {
		const state = createState();
		render(<AnalyticsOverviewScreen state={state} />);

		expect(screen.getByText('abc123def456'.slice(0, 12) + '...')).toBeInTheDocument();
	});

	it('renders health section with status chip and no errors', () => {
		const state = createState();
		render(<AnalyticsOverviewScreen state={state} />);

		expect(screen.getAllByText('Healthy').length).toBeGreaterThan(0);
		expect(screen.getByText('ANALYTICS_NO_ERRORS')).toBeInTheDocument();
	});

	it('renders health section with recent errors', () => {
		const overview = createOverview({
			health: {
				status: 'error',
				last_scan_at: '2026-03-16T11:40:00Z',
				last_scan_seconds: 12,
				indexed_files: 200,
				errors_last_24h: 2,
				recent_errors: ['disk failure', 'timeout'],
			},
		});
		const state = createState({ data: overview, healthStatusLabel: 'Error', healthStatusColor: 'error' });
		render(<AnalyticsOverviewScreen state={state} />);

		expect(screen.getByText('disk failure')).toBeInTheDocument();
		expect(screen.getByText('timeout')).toBeInTheDocument();
	});

	it('renders used percent and progress', () => {
		const state = createState({ usedPercent: 40 });
		render(<AnalyticsOverviewScreen state={state} />);

		expect(screen.getByRole('progressbar')).toBeInTheDocument();
	});

	it('renders action buttons that navigate', () => {
		const state = createState();
		render(<AnalyticsOverviewScreen state={state} />);

		fireEvent.click(screen.getByText('ANALYTICS_ACTION_VIEW_DUPLICATES'));
		expect(mockNavigate).toHaveBeenCalled();

		fireEvent.click(screen.getByText('ANALYTICS_ACTION_VIEW_RECENT'));
		expect(mockNavigate).toHaveBeenCalledTimes(2);

		fireEvent.click(screen.getByText('ANALYTICS_ACTION_VIEW_LARGEST'));
		expect(mockNavigate).toHaveBeenCalledTimes(3);
	});

	it('calls refresh on reindex button click', () => {
		const state = createState();
		render(<AnalyticsOverviewScreen state={state} />);

		fireEvent.click(screen.getByText('ANALYTICS_ACTION_REINDEX'));
		expect(state.refresh).toHaveBeenCalled();
	});

	it('handles null data gracefully', () => {
		const state = createState({ data: null });
		render(<AnalyticsOverviewScreen state={state} />);

		expect(screen.getByText('ANALYTICS_KPI_STORAGE')).toBeInTheDocument();
		expect(screen.getByText('0 B / 0 B')).toBeInTheDocument();
		expect(screen.getByText('ANALYTICS_KPI_FILES_ADDED')).toBeInTheDocument();
	});

	it('handles empty hot_folders', () => {
		const overview = createOverview({ hot_folders: [] });
		const state = createState({ data: overview });
		render(<AnalyticsOverviewScreen state={state} />);

		expect(screen.getByText('0')).toBeInTheDocument();
	});

	it('caps progress bar at 100', () => {
		const state = createState({ usedPercent: 120 });
		render(<AnalyticsOverviewScreen state={state} />);

		const progressbar = screen.getByRole('progressbar');
		expect(progressbar.getAttribute('aria-valuenow')).toBe('100');
	});
});
