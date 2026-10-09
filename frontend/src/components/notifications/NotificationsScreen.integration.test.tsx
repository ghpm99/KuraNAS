import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter, useLocation } from 'react-router-dom';
import NotificationsScreen from './NotificationsScreen';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), put: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; put: jest.Mock };

const buildNotification = (id: number, overrides: Record<string, unknown> = {}) => ({
	id,
	type: 'info',
	title: `Notificacao ${id}`,
	message: 'ok',
	is_read: false,
	created_at: '2026-06-01T00:00:00Z',
	group_count: 1,
	is_grouped: false,
	...overrides,
});

const buildPage = (items: unknown[], page = 1, hasNext = false) => ({
	items,
	pagination: { page, page_size: 20, has_next: hasNext, has_prev: page > 1 },
});

const LocationProbe = () => <span data-testid="location">{useLocation().pathname}</span>;

const renderScreen = () => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={client}>
			<MemoryRouter>
				<NotificationsScreen />
				<LocationProbe />
			</MemoryRouter>
		</QueryClientProvider>
	);
};

describe('components/notifications/NotificationsScreen (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockResolvedValue({ data: buildPage([buildNotification(9, { title: 'Backup terminou' })]) });
		mockedApi.put.mockResolvedValue({ data: undefined });
	});

	it('renders the page header and an error state with retry when the backend is absent', async () => {
		mockedApi.get.mockRejectedValueOnce({ response: { data: { error: 'falha no servidor' } } });

		renderScreen();

		expect(screen.getByRole('heading', { level: 1, name: 'NOTIFICATIONS' })).toBeInTheDocument();
		expect(await screen.findByText('falha no servidor')).toBeInTheDocument();
		expect(screen.getByText('NOTIFICATIONS_LOAD_ERROR')).toBeInTheDocument();

		fireEvent.click(screen.getByRole('button', { name: 'TRY_AGAIN' }));

		expect(await screen.findByText('Backup terminou')).toBeInTheDocument();
	});

	it('shows the shared empty state when there are no notifications', async () => {
		mockedApi.get.mockResolvedValue({ data: buildPage([]) });

		renderScreen();

		expect(await screen.findByRole('status')).toHaveTextContent('NO_NOTIFICATIONS');
	});

	it('clicking an unread item issues PUT /notifications/:id/read', async () => {
		renderScreen();
		await screen.findByText('Backup terminou');

		fireEvent.click(screen.getByRole('button', { name: /Backup terminou/ }));

		await waitFor(() => expect(mockedApi.put).toHaveBeenCalledWith('/notifications/9/read'));
		expect(screen.getByTestId('location')).toHaveTextContent('/');
	});

	it('clicking a notification with a target marks it read and navigates there', async () => {
		mockedApi.get.mockResolvedValue({
			data: buildPage([
				buildNotification(4, { title: 'Captura pronta', metadata: { event: 'capture_promoted' } }),
			]),
		});

		renderScreen();
		await screen.findByText('Captura pronta');

		fireEvent.click(screen.getByRole('button', { name: /Captura pronta/ }));

		await waitFor(() => expect(mockedApi.put).toHaveBeenCalledWith('/notifications/4/read'));
		expect(screen.getByTestId('location')).toHaveTextContent('/captures');
	});

	it('mark-all issues PUT /notifications/read-all', async () => {
		renderScreen();
		await screen.findByText('Backup terminou');

		fireEvent.click(screen.getByText('MARK_ALL_AS_READ'));

		await waitFor(() => expect(mockedApi.put).toHaveBeenCalledWith('/notifications/read-all'));
	});

	it('loads the next page through the shared load-more sentinel', async () => {
		mockedApi.get.mockImplementation((_url: string, config: { params: { page: number } }) =>
			Promise.resolve({
				data: buildPage(
					[buildNotification(config.params.page, { title: `Pagina ${config.params.page}` })],
					config.params.page,
					config.params.page === 1
				),
			})
		);

		renderScreen();
		await screen.findByText('Pagina 1');

		fireEvent.click(screen.getByRole('button', { name: 'LOAD_MORE' }));

		expect(await screen.findByText('Pagina 2')).toBeInTheDocument();
		expect(mockedApi.get).toHaveBeenCalledWith('/notifications', {
			params: expect.objectContaining({ page: 2, page_size: 20 }),
		});
	});

	it('requests only unread notifications when the unread filter is selected', async () => {
		renderScreen();
		await screen.findByText('Backup terminou');

		fireEvent.click(screen.getByRole('tab', { name: 'UNREAD' }));

		await waitFor(() =>
			expect(mockedApi.get).toHaveBeenCalledWith('/notifications', {
				params: expect.objectContaining({ is_read: false }),
			})
		);
	});
});
