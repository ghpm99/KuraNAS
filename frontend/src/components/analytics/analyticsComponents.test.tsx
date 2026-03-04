import { render, screen } from '@testing-library/react';
import React from 'react';
import ActivityChart from './ActivityChart/ActivityChart';
import BackupSection from './BackupSection/BackupSection';
import CleanupSuggestions from './CleanupSuggestions/CleanupSuggestions';
import DiskUsageChart from './DiskUsageChart/DiskUsageChart';
import DuplicatesSection from './DuplicatesSection/DuplicatesSection';
import EmptyFoldersSection from './EmptyFoldersSection/EmptyFoldersSection';
import FileTypesChart from './FileTypesChart/FileTypesChart';
import FileTypesTable from './FileTypesTable/FileTypesTable';
import LargestFilesTable from './LargestFilesTable/LargestFilesTable';
import RecentActivity from './RecentActivity/RecentActivity';
import SizeRangesChart from './SizeRangesChart/SizeRangesChart';
import StorageOverviewCards from './StorageOverviewCards/StorageOverviewCards';
import TrashSection from './TrashSection/TrashSection';
import AnalyticsLayout from './analyticsLayout';

jest.mock('@/components/contexts/AnalyticsContext', () => ({
	useAnalytics: jest.fn(),
	AnalyticsProvider: ({ children }: any) => <div data-testid='analytics-provider'>{children}</div>,
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (k: string) => k }),
}));

jest.mock('@/components/ui/Card/Card', () => ({ title, children }: any) => (
	<div>
		<h3>{title}</h3>
		{children}
	</div>
));

jest.mock('@mui/x-charts', () => ({
	PieChart: () => <div data-testid='pie-chart' />,
}));

jest.mock('../layout/Layout', () => ({ children }: any) => <div data-testid='layout'>{children}</div>);

const { useAnalytics } = jest.requireMock('@/components/contexts/AnalyticsContext');

const baseData = {
	storageOverview: {
		totalUsedSpace: '100 GB',
		totalFiles: 10,
		totalFolders: 5,
		availableSpace: '900 GB',
		diskUsage: { used: 60, free: 40 },
	},
	fileTypes: [
		{ format: '.mp3', total: 3, size: 1024, percentage: 30 },
		{ format: null, total: 1, size: 512, percentage: 10 },
	],
	sizeRanges: [{ range: '<10MB', count: 5 }],
	largestFiles: [{ name: 'movie.mkv', size: 2048, path: '/videos' }],
	duplicates: { total: 1, total_size: 1000, files: [{ name: 'dup.txt', size: 1000, copies: 2, paths: ['/a', '/b', '/c'] }] },
	recentActivity: {
		recentFiles: [{ name: 'new.txt', size: '1 KB', date: '2026-01-01' }],
		accessedFiles: [{ name: 'seen.txt', accessCount: 4, lastAccess: '2026-01-01' }],
		activityChart: [{ date: '2026-01-01', created: 2, modified: 3 }],
	},
	organization: { emptyFolders: 2, emptyPaths: ['/empty'] },
	cleanup: {
		oldLargeFiles: [{ name: 'old.zip', size: '2 GB', path: '/old' }],
		similarNames: [{ name1: 'a', name2: 'b', similarity: 90 }],
		criticalSpace: true,
	},
	backup: {
		lastBackup: '2026-01-01',
		lastBackupSize: '1 TB',
		history: [{ date: '2026-01-01', size: '1 TB', status: 'success' }],
	},
	trash: {
		totalFiles: 1,
		totalSpace: '10 MB',
		files: [{ name: 'trash.txt', size: '10 MB', deletedDate: '2026-01-01' }],
	},
};

describe('analytics components', () => {
	beforeEach(() => {
		useAnalytics.mockReturnValue({ analyticsData: baseData });
	});

	it('renders analytics widgets', () => {
		render(
			<>
				<ActivityChart />
				<BackupSection />
				<CleanupSuggestions />
				<DiskUsageChart />
				<DuplicatesSection />
				<EmptyFoldersSection />
				<FileTypesChart />
				<FileTypesTable />
				<LargestFilesTable />
				<RecentActivity />
				<SizeRangesChart />
				<StorageOverviewCards />
				<TrashSection />
			</>,
		);

		expect(screen.getByText('ANALYTICS_BACKUP_HISTORY')).toBeInTheDocument();
		expect(screen.getByText('dup.txt')).toBeInTheDocument();
		expect(screen.getAllByTestId('pie-chart').length).toBeGreaterThan(0);
		expect(screen.getByText('ANALYTICS_CRITICAL_SPACE')).toBeInTheDocument();
	});

	it('handles no duplicates branch', () => {
		useAnalytics.mockReturnValue({
			analyticsData: {
				...baseData,
				duplicates: { total: 0, total_size: 0, files: [] },
			},
		});

		render(<DuplicatesSection />);
		expect(screen.getByText('ANALYTICS_NO_DUPLICATES')).toBeInTheDocument();
	});

	it('handles backup fallback status and cleanup without alert', () => {
		useAnalytics.mockReturnValue({
			analyticsData: {
				...baseData,
				backup: { ...baseData.backup, history: [{ date: 'x', size: 'y', status: 'unknown' }] },
				cleanup: { ...baseData.cleanup, criticalSpace: false },
			},
		});

		render(
			<>
				<BackupSection />
				<CleanupSuggestions />
			</>,
		);

		expect(screen.getByText('ANALYTICS_BACKUP_PENDING')).toBeInTheDocument();
		expect(screen.queryByText('ANALYTICS_CRITICAL_SPACE')).not.toBeInTheDocument();
	});

	it('renders analytics layout wrappers', () => {
		render(
			<AnalyticsLayout>
				<div>body</div>
			</AnalyticsLayout>,
		);
		expect(screen.getByTestId('analytics-provider')).toBeInTheDocument();
		expect(screen.getByTestId('layout')).toBeInTheDocument();
		expect(screen.getByText('body')).toBeInTheDocument();
	});
});
