import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import NotificationsScreen from './NotificationsScreen';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real NotificationsScreen +
// service/notifications.ts run, so each command asserts the exact PUT endpoint
// the backend notifications handlers expose.
jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), put: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; put: jest.Mock };

const page = {
	items: [
		{ id: 9, type: 'info', title: 'Backup terminou', message: 'ok', is_read: false, created_at: '2026-06-01T00:00:00Z' },
	],
	pagination: { page: 1, page_size: 20, has_next: false, has_prev: false },
};

const renderScreen = () => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={client}>
			<NotificationsScreen />
		</QueryClientProvider>
	);
};

describe('components/notifications/NotificationsScreen (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockResolvedValue({ data: page });
		mockedApi.put.mockResolvedValue({ data: undefined });
	});

	it('clicking an unread item issues PUT /notifications/:id/read', async () => {
		renderScreen();
		await screen.findByText('Backup terminou');

		fireEvent.click(screen.getByText('Backup terminou'));

		await waitFor(() => expect(mockedApi.put).toHaveBeenCalledWith('/notifications/9/read'));
	});

	it('mark-all issues PUT /notifications/read-all', async () => {
		renderScreen();
		await screen.findByText('Backup terminou');

		fireEvent.click(screen.getByText('MARK_ALL_AS_READ'));

		await waitFor(() => expect(mockedApi.put).toHaveBeenCalledWith('/notifications/read-all'));
	});
});
