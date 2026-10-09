import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import NotificationBell from './NotificationBell';
import NotificationProvider from '@/components/providers/notificationProvider';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), put: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; put: jest.Mock };

const renderBell = () => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={client}>
			<MemoryRouter>
				<NotificationProvider>
					<NotificationBell />
				</NotificationProvider>
			</MemoryRouter>
		</QueryClientProvider>
	);
};

const notificationsPage = {
	items: [
		{
			id: 1,
			type: 'info',
			title: 'Fresh notification',
			message: 'body',
			is_read: true,
			created_at: '2026-06-01T00:00:00Z',
			group_count: 1,
			is_grouped: false,
		},
	],
	pagination: { page: 1, page_size: 5, has_next: false, has_prev: false },
};

describe('components/notifications/NotificationBell', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('renders and opens with an absent backend without crashing', async () => {
		mockedApi.get.mockRejectedValue(new Error('offline'));

		renderBell();
		fireEvent.click(screen.getByRole('button', { name: 'NOTIFICATIONS' }));

		expect(await screen.findByText('NO_NOTIFICATIONS')).toBeInTheDocument();
	});

	it('refetches the notification list every time the popover opens', async () => {
		mockedApi.get.mockImplementation((url: string) =>
			Promise.resolve({ data: url === '/notifications/unread-count' ? { unread_count: 0 } : notificationsPage })
		);

		renderBell();
		await waitFor(() => expect(mockedApi.get).toHaveBeenCalledWith('/notifications', expect.anything()));
		const listCallsBeforeOpen = mockedApi.get.mock.calls.filter(([url]) => url === '/notifications').length;

		fireEvent.click(screen.getByRole('button', { name: 'NOTIFICATIONS' }));

		await screen.findByText('Fresh notification');
		await waitFor(() => {
			const listCalls = mockedApi.get.mock.calls.filter(([url]) => url === '/notifications').length;
			expect(listCalls).toBeGreaterThan(listCallsBeforeOpen);
		});
	});
});
