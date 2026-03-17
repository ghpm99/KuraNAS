import { act, renderHook } from '@testing-library/react';
import { useAnalyticsScreenState } from './useAnalyticsScreenState';
import type { AnalyticsOverview } from '@/types/analytics';

const mockUseAnalyticsOverview = jest.fn();

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string, params?: Record<string, string | number>) => {
			if (key === 'ANALYTICS_UPDATED_MINUTES') return `Updated ${params?.minutes ?? '-'}`;
			const map: Record<string, string> = {
				ANALYTICS_STATUS_OK: 'Healthy',
				ANALYTICS_STATUS_SCANNING: 'Scanning',
				ANALYTICS_STATUS_ERROR: 'Error',
			};
			return map[key] ?? key;
		},
	}),
}));

jest.mock('@/components/providers/analyticsProvider/analyticsContext', () => ({
	useAnalyticsOverview: () => mockUseAnalyticsOverview(),
}));

const createOverview = (overrides?: Partial<AnalyticsOverview>): AnalyticsOverview => ({
	period: '7d',
	generated_at: '2026-03-16T11:50:00Z',
	storage: {
		total_bytes: 200,
		used_bytes: 50,
		free_bytes: 150,
		growth_bytes: 10,
	},
	counts: {
		files_total: 20,
		files_added: 2,
		folders: 3,
	},
	time_series: [],
	types: [],
	extensions: [],
	hot_folders: [],
	top_folders: [],
	recent_files: [],
	duplicates: {
		groups: 1,
		files: 2,
		reclaimable_size: 20,
		top_groups: [],
	},
	library: {
		categorized_media: 1,
		audio_with_metadata: 1,
		video_with_metadata: 1,
		image_with_metadata: 1,
		image_classified: 1,
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
		indexed_files: 100,
		errors_last_24h: 0,
		recent_errors: [],
	},
	...overrides,
});

describe('analytics/useAnalyticsScreenState', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		jest.useFakeTimers();
		jest.setSystemTime(new Date('2026-03-16T12:00:00Z'));
	});

	afterEach(() => {
		jest.useRealTimers();
	});

	it('returns safe defaults when overview data is missing', () => {
		mockUseAnalyticsOverview.mockReturnValue({
			period: '24h',
			setPeriod: jest.fn(),
			data: null,
			loading: false,
			error: null,
			refresh: jest.fn(),
		});

		const { result, unmount } = renderHook(() => useAnalyticsScreenState());

		expect(result.current.updatedMinutes).toBe('-');
		expect(result.current.healthStatusLabel).toBe('Healthy');
		expect(result.current.healthStatusColor).toBe('success');
		expect(result.current.usedPercent).toBe(0);
		expect(result.current.reclaimablePercent).toBe(0);
		expect(result.current.processingFailureTotal).toBe(0);
		expect(result.current.formatBytes(Number.NaN)).toBe('0 B');
		expect(result.current.formatBytes(12)).toBe('12 B');
		expect(result.current.formatDate('')).toBe('-');
		expect(result.current.formatDate('invalid-date')).toBe('-');
		expect(result.current.formatPercent(12.34)).toBe('12.3%');

		unmount();
	});

	it('derives status, formatting and elapsed minutes from analytics data', () => {
		const setPeriod = jest.fn();
		const refresh = jest.fn();
		const currentValue = {
			period: '7d',
			setPeriod,
			data: createOverview(),
			loading: false,
			error: null,
			refresh,
		};

		mockUseAnalyticsOverview.mockImplementation(() => currentValue);

		const { result, rerender } = renderHook(() => useAnalyticsScreenState());

		expect(result.current.updatedMinutes).toBe('10');
		expect(result.current.healthStatusLabel).toBe('Healthy');
		expect(result.current.healthStatusColor).toBe('success');
		expect(result.current.usedPercent).toBe(25);
		expect(result.current.reclaimablePercent).toBe(10);
		expect(result.current.processingFailureTotal).toBe(3);
		expect(result.current.formatBytes(1536)).toBe('1.5 KB');
		expect(result.current.formatBytes(100 * 1024 * 1024)).toBe('100 MB');
		expect(result.current.formatDate('2026-03-16T11:30:00Z')).not.toBe('-');

		act(() => {
			jest.advanceTimersByTime(60000);
		});
		expect(result.current.updatedMinutes).toBe('11');

		currentValue.data = createOverview({
			generated_at: 'invalid-date',
			storage: {
				total_bytes: 0,
				used_bytes: 0,
				free_bytes: 0,
				growth_bytes: 0,
			},
			duplicates: {
				groups: 0,
				files: 0,
				reclaimable_size: 0,
				top_groups: [],
			},
			health: {
				status: 'scanning',
				last_scan_at: '',
				last_scan_seconds: 0,
				indexed_files: 0,
				errors_last_24h: 0,
				recent_errors: [],
			},
			processing: {
				metadata_pending: 0,
				metadata_failed: 0,
				thumbnail_pending: 0,
				thumbnail_failed: 0,
			},
		});
		rerender();

		expect(result.current.updatedMinutes).toBe('-');
		expect(result.current.healthStatusLabel).toBe('Scanning');
		expect(result.current.healthStatusColor).toBe('warning');
		expect(result.current.usedPercent).toBe(0);
		expect(result.current.reclaimablePercent).toBe(0);

		currentValue.data = createOverview({
			health: {
				status: 'error',
				last_scan_at: '',
				last_scan_seconds: 0,
				indexed_files: 0,
				errors_last_24h: 3,
				recent_errors: ['disk error'],
			},
		});
		rerender();

		expect(result.current.healthStatusLabel).toBe('Error');
		expect(result.current.healthStatusColor).toBe('error');
		expect(result.current.period).toBe('7d');
		expect(result.current.setPeriod).toBe(setPeriod);
		expect(result.current.refresh).toBe(refresh);
	});
});
