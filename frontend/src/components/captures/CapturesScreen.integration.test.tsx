import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import CapturesScreen from './CapturesScreen';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real CapturesScreen + useCapturesScreen
// + service/captures.ts run, so the delete button asserts DELETE /captures/:id
// the backend captures handler decodes.
jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), delete: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; delete: jest.Mock };

const page = {
	items: [
		{
			id: 7,
			name: 'recording_1',
			file_name: 'recording_1.mp4',
			file_path: '/x/recording_1.mp4',
			media_type: 'video',
			mime_type: 'video/mp4',
			size: 5 * 1024 * 1024,
			episode_key: '',
			created_at: '2026-06-01T00:00:00Z',
			status: 'uploaded',
		},
	],
	pagination: { page: 1, page_size: 20, has_next: false, has_prev: false },
};

const renderScreen = () => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={client}>
			<CapturesScreen />
		</QueryClientProvider>
	);
};

describe('CapturesScreen (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockResolvedValue({ data: page });
		mockedApi.delete.mockResolvedValue({ data: undefined });
	});

	it('deleting a capture issues DELETE /captures/:id', async () => {
		renderScreen();
		await screen.findByText('recording_1');

		fireEvent.click(screen.getByLabelText('CAPTURES_DELETE'));

		await waitFor(() => expect(mockedApi.delete).toHaveBeenCalledWith('/captures/7'));
	});
});
