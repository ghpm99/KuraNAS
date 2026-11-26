import { apiBase } from '@/service';
import type { TakeoutImport } from '@/types/takeout';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import TakeoutImportSummary from './TakeoutImportSummary';

jest.mock('@/service', () => ({
	apiBase: { get: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string, variables?: Record<string, string>) =>
			variables ? `${key}=${variables.count}` : key,
	}),
}));

const completedImport: TakeoutImport = {
	id: 7,
	job_id: 3,
	file_name: 'takeout.zip',
	size: 100,
	state: 'completed',
	total_entries: 10,
	processed_entries: 10,
	imported: 6,
	already_existing: 2,
	ignored: 1,
	failed_entries: 1,
	error: '',
	created_at: '2026-11-25T10:00:00Z',
	ended_at: null,
};

describe('TakeoutImportSummary counts', () => {
	it('shows the four counts of the finished import', () => {
		(apiBase.get as jest.Mock).mockRejectedValue(new Error('offline'));

		render(
			<QueryClientProvider
				client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
			>
				<MemoryRouter>
					<TakeoutImportSummary takeoutImport={completedImport} />
				</MemoryRouter>
			</QueryClientProvider>
		);

		expect(screen.getByText('TAKEOUT_SUMMARY_IMPORTED=6')).toBeInTheDocument();
		expect(screen.getByText('TAKEOUT_SUMMARY_ALREADY_EXISTING=2')).toBeInTheDocument();
		expect(screen.getByText('TAKEOUT_SUMMARY_IGNORED=1')).toBeInTheDocument();
		expect(screen.getByText('TAKEOUT_SUMMARY_FAILED=1')).toBeInTheDocument();
	});
});
