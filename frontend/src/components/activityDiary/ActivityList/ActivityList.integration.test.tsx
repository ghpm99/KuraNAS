import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import ActivityDiaryProvider from '@/components/providers/activityDiaryProvider';
import ActivityList from './ActivityList';
import { apiBase } from '@/service';

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
});
