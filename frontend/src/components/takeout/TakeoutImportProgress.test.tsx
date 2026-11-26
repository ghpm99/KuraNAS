import { act, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { apiBase } from '@/service';
import TakeoutImportProgress from './TakeoutImportProgress';

jest.mock('@/service', () => ({
	apiBase: { get: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string, variables?: Record<string, unknown>) =>
			variables ? `${key} ${variables.processed}/${variables.total}` : key,
	}),
}));

const renderProgress = () =>
	render(
		<QueryClientProvider
			client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
		>
			<MemoryRouter>
				<TakeoutImportProgress importId={7} />
			</MemoryRouter>
		</QueryClientProvider>
	);

const mockedApi = apiBase as unknown as { get: jest.Mock };

const importPollCalls = () =>
	mockedApi.get.mock.calls.filter(([url]) => url === '/takeout/imports/7');

const buildImport = (state: string, processedEntries: number) => ({
	id: 7,
	job_id: 3,
	file_name: 'takeout.zip',
	size: 100,
	state,
	total_entries: 10,
	processed_entries: processedEntries,
	imported: 0,
	already_existing: 0,
	ignored: 0,
	failed_entries: 0,
	error: '',
	created_at: '2026-11-25T10:00:00Z',
	ended_at: null,
});

describe('TakeoutImportProgress', () => {
	beforeEach(() => {
		jest.useFakeTimers({ advanceTimers: true });
		jest.clearAllMocks();
	});

	afterEach(() => {
		jest.useRealTimers();
	});

	it('polls every 2s while running, shows processed/total and stops when completed', async () => {
		mockedApi.get
			.mockResolvedValueOnce({ data: buildImport('running', 3) })
			.mockResolvedValueOnce({ data: buildImport('running', 6) })
			.mockResolvedValue({ data: buildImport('completed', 10) });

		const { container } = renderProgress();

		expect(await screen.findByText('TAKEOUT_IMPORT_PROGRESS 3/10')).toBeInTheDocument();
		expect(mockedApi.get).toHaveBeenCalledWith('/takeout/imports/7');
		expect(container.textContent).not.toContain('job #');

		await act(async () => {
			await jest.advanceTimersByTimeAsync(2000);
		});
		await waitFor(() => expect(importPollCalls()).toHaveLength(2));

		await act(async () => {
			await jest.advanceTimersByTimeAsync(2000);
		});
		await waitFor(() => expect(importPollCalls()).toHaveLength(3));

		await act(async () => {
			await jest.advanceTimersByTimeAsync(10000);
		});
		expect(importPollCalls()).toHaveLength(3);
		expect(await screen.findByText('TAKEOUT_IMPORT_FINISHED')).toBeInTheDocument();
	});

	it('shows the summary and the shortcuts to the imported photos and videos when completed', async () => {
		mockedApi.get.mockImplementation(async (url: string) => {
			if (url === '/takeout/imports/7') return { data: buildImport('completed', 10) };
			if (url === '/libraries') {
				return {
					data: [
						{ category: 'images', path: '/nas/Imagens' },
						{ category: 'videos', path: '/nas/Videos' },
					],
				};
			}
			return { data: { path: url === '/files/by-disk-path' ? '/Pasta' : '' } };
		});

		renderProgress();

		expect(await screen.findByText('TAKEOUT_IMPORT_FINISHED')).toBeInTheDocument();
		expect(screen.getByText(/TAKEOUT_SUMMARY_IMPORTED/)).toBeInTheDocument();
		expect(
			screen.getByRole('link', { name: 'TAKEOUT_VIEW_IMPORTED_PHOTOS' })
		).toBeInTheDocument();
		expect(
			screen.getByRole('link', { name: 'TAKEOUT_VIEW_IMPORTED_VIDEOS' })
		).toBeInTheDocument();
	});

	it('shows the server error message when polling fails', async () => {
		mockedApi.get.mockRejectedValue({
			response: { data: { error: 'importação não encontrada' } },
		});

		renderProgress();

		expect(await screen.findByText('importação não encontrada')).toBeInTheDocument();
	});
});
