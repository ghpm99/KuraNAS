import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import ActivityDiaryProvider from '@/components/providers/activityDiaryProvider';
import ActivityDiaryForm from './ActivityDiaryForm';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real ActivityDiaryProvider +
// ActivityDiaryForm + service/activityDiary.ts run, so submitting asserts POST
// /diary/ with the typed name/description the backend diary handler decodes.
jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), post: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock };

const renderForm = () => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={client}>
			<ActivityDiaryProvider>
				<ActivityDiaryForm />
			</ActivityDiaryProvider>
		</QueryClientProvider>
	);
};

describe('components/activityDiary/ActivityDiaryForm (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockImplementation((url: string) => {
			if (url === '/diary/summary') return Promise.resolve({ data: { total: 0, longest: null } });
			return Promise.resolve({ data: { items: [], pagination: { has_next: false } } });
		});
		mockedApi.post.mockResolvedValue({ data: { ID: 1, name: 'Treino', description: 'Corrida' } });
	});

	it('submitting the form issues POST /diary/ with the typed name and description', async () => {
		renderForm();

		// The form lowercases the name field as the user types.
		fireEvent.change(screen.getByPlaceholderText('ACTIVITY_NAME_PLACEHOLDER'), {
			target: { value: 'treino' },
		});
		fireEvent.change(screen.getByPlaceholderText('ACTIVITY_DESCRIPTION_PLACEHOLDER'), {
			target: { value: 'Corrida' },
		});
		fireEvent.click(screen.getByText('ADD_ACTIVITY'));

		await waitFor(() =>
			expect(mockedApi.post).toHaveBeenCalledWith('/diary/', { name: 'treino', description: 'Corrida' })
		);
	});
});
