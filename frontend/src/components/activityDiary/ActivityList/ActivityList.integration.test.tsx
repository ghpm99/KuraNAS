import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import ActivityDiaryProvider from '@/components/providers/activityDiaryProvider';
import ActivityList from './ActivityList';
import { apiBase } from '@/service';
import { buildServerPage } from '@/shared/test/buildServerPage';

// Seam test: only apiBase is mocked; the real ActivityDiaryProvider +
// ActivityList + service/activityDiary.ts run, so clicking the copy button
// asserts POST /diary/copy with the id key the backend DiaryId decodes.
jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), post: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock };

const activity = {
	id: 5,
	name: 'Treino',
	description: 'Corrida',
	start_time: '2026-06-01T10:00:00Z',
	end_time: { HasValue: true, Value: '2026-06-01T11:00:00Z' },
	duration: 3600,
};

const renderList = () => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={client}>
			<ActivityDiaryProvider>
				<ActivityList />
			</ActivityDiaryProvider>
		</QueryClientProvider>
	);
};

describe('components/activityDiary/ActivityList (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockImplementation((url: string) => {
			if (url === '/diary/summary') return Promise.resolve({ data: { total: 1, longest: null } });
			return Promise.resolve({
				data: { items: [activity], pagination: { has_next: false } },
			});
		});
		mockedApi.post.mockResolvedValue({ data: { ...activity, id: 6 } });
	});

	it('clicking copy issues POST /diary/copy with the entry id', async () => {
		renderList();
		await screen.findByText('Treino');

		fireEvent.click(screen.getByRole('button'));

		await waitFor(() => expect(mockedApi.post).toHaveBeenCalledWith('/diary/copy', { id: 5 }));
	});

	describe('page failures', () => {
		const firstPageActivity = { ...activity, id: 1, name: 'Primeira' };
		const secondPageActivity = { ...activity, id: 2, name: 'Segunda' };
		const diaryPageSize = 20;

		const summaryResponse = { data: { total: 1, longest: null } };

		it('keeps first page entries when page 2 fails and retries only page 2', async () => {
			let secondPageAttempts = 0;
			mockedApi.get.mockImplementation((url: string, config?: { params?: { page: number } }) => {
				if (url === '/diary/summary') return Promise.resolve(summaryResponse);
				if (config?.params?.page === 2) {
					secondPageAttempts += 1;
					if (secondPageAttempts === 1) return Promise.reject(new Error('page 2 down'));
					return Promise.resolve({
						data: buildServerPage([secondPageActivity], {
							page: 2,
							pageSize: diaryPageSize,
							hasNext: false,
						}),
					});
				}
				return Promise.resolve({
					data: buildServerPage([firstPageActivity], {
						page: 1,
						pageSize: diaryPageSize,
						hasNext: true,
					}),
				});
			});

			renderList();
			await screen.findByText('Primeira');
			fireEvent.click(await screen.findByRole('button', { name: 'LOAD_MORE' }));

			await screen.findByText('ACTIVITY_DIARY_LOAD_ERROR');
			expect(screen.getByText('Primeira')).toBeInTheDocument();

			mockedApi.get.mockClear();
			fireEvent.click(screen.getByRole('button', { name: 'TRY_AGAIN' }));

			await screen.findByText('Segunda');
			expect(screen.getByText('Primeira')).toBeInTheDocument();
			const requestedPages = mockedApi.get.mock.calls
				.filter(([url]: [string]) => url === '/diary/')
				.map(([, config]: [string, { params: { page: number } }]) => config.params.page);
			expect(requestedPages).toEqual([2]);
		});

		it('shows the error state, not the empty one, when the server is down from the start', async () => {
			mockedApi.get.mockImplementation((url: string) => {
				if (url === '/diary/summary') return Promise.resolve(summaryResponse);
				return Promise.reject(new Error('server down'));
			});

			renderList();

			await screen.findByText('ACTIVITY_DIARY_LOAD_ERROR');
			expect(screen.queryByText('NO_ACTIVITIES')).not.toBeInTheDocument();
		});
	});
});
