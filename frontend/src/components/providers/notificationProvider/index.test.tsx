import { act, render, screen } from '@testing-library/react';
import { useNotifications } from './notificationContext';
import { NotificationProvider } from './index';
import {
	getNotifications,
	getUnreadCount,
	markAllNotificationsAsRead,
	markNotificationAsRead,
} from '@/service/notifications';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

jest.mock('@/service/notifications', () => ({
	getNotifications: jest.fn(),
	getUnreadCount: jest.fn(),
	markNotificationAsRead: jest.fn(),
	markAllNotificationsAsRead: jest.fn(),
}));

jest.mock('@tanstack/react-query', () => ({
	useQuery: jest.fn(),
	useMutation: jest.fn(),
	useQueryClient: jest.fn(),
}));

const mockedUseQuery = useQuery as jest.Mock;
const mockedUseMutation = useMutation as jest.Mock;
const mockedUseQueryClient = useQueryClient as jest.Mock;

const mockedGetNotifications = getNotifications as jest.Mock;
const mockedGetUnreadCount = getUnreadCount as jest.Mock;
const mockedMarkNotificationAsRead = markNotificationAsRead as jest.Mock;
const mockedMarkAllNotificationsAsRead = markAllNotificationsAsRead as jest.Mock;

function Consumer() {
	const value = useNotifications();

	return (
		<div>
			<span data-testid="unread-count">{String(value.unreadCount)}</span>
			<span data-testid="notifications-size">{String(value.notifications.length)}</span>
			<span data-testid="is-loading">{String(value.isLoading)}</span>
			<button type="button" onClick={() => value.markAsRead(7)}>
				mark-one
			</button>
			<button type="button" onClick={() => value.markAllAsRead()}>
				mark-all
			</button>
			<button type="button" onClick={() => value.refetch()}>
				refetch
			</button>
		</div>
	);
}

describe('components/providers/notificationProvider/index', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('exposes default values when queries return empty data', () => {
		mockedUseQueryClient.mockReturnValue({ invalidateQueries: jest.fn() });
		mockedUseQuery.mockReturnValue({
			data: undefined,
			isLoading: false,
			refetch: jest.fn().mockResolvedValue(undefined),
		});
		mockedUseMutation.mockReturnValue({ mutateAsync: jest.fn() });

		render(
			<NotificationProvider>
				<Consumer />
			</NotificationProvider>
		);

		expect(screen.getByTestId('unread-count')).toHaveTextContent('0');
		expect(screen.getByTestId('notifications-size')).toHaveTextContent('0');
		expect(screen.getByTestId('is-loading')).toHaveTextContent('false');
	});

	it('wires mark/refetch actions and invalidates notification caches on mutation success', async () => {
		const invalidateQueries = jest.fn().mockResolvedValue(undefined);
		const unreadRefetch = jest.fn().mockResolvedValue(undefined);
		const listRefetch = jest.fn().mockResolvedValue(undefined);
		const mutateOne = jest.fn().mockResolvedValue(undefined);
		const mutateAll = jest.fn().mockResolvedValue(undefined);
		const mutationOptions: Array<{ onSuccess?: () => Promise<void> }> = [];

		mockedUseQueryClient.mockReturnValue({ invalidateQueries });
		mockedUseQuery.mockImplementation((options: { queryKey: string[] }) => {
			if (options.queryKey[0] === 'notifications-unread-count') {
				return {
					data: { unread_count: 3 },
					isLoading: false,
					refetch: unreadRefetch,
				};
			}
			return {
				data: { items: [{ id: 1 }] },
				isLoading: true,
				refetch: listRefetch,
			};
		});
		mockedUseMutation.mockImplementation((options: { mutationFn: unknown; onSuccess?: () => Promise<void> }) => {
			mutationOptions.push(options);
			if (options.mutationFn === markNotificationAsRead) {
				return { mutateAsync: mutateOne };
			}
			return { mutateAsync: mutateAll };
		});

		render(
			<NotificationProvider>
				<Consumer />
			</NotificationProvider>
		);

		expect(screen.getByTestId('unread-count')).toHaveTextContent('3');
		expect(screen.getByTestId('notifications-size')).toHaveTextContent('1');
		expect(screen.getByTestId('is-loading')).toHaveTextContent('true');

		await act(async () => {
			screen.getByRole('button', { name: 'mark-one' }).click();
		});
		expect(mutateOne).toHaveBeenCalledWith(7);

		await act(async () => {
			screen.getByRole('button', { name: 'mark-all' }).click();
		});
		expect(mutateAll).toHaveBeenCalled();

		await act(async () => {
			screen.getByRole('button', { name: 'refetch' }).click();
		});
		expect(unreadRefetch).toHaveBeenCalled();
		expect(listRefetch).toHaveBeenCalled();

		await act(async () => {
			await mutationOptions[0]?.onSuccess?.();
			await mutationOptions[1]?.onSuccess?.();
		});

		expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['notifications-unread-count'] });
		expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['notifications'] });
	});

	it('uses service functions as query and mutation functions', async () => {
		mockedUseQueryClient.mockReturnValue({ invalidateQueries: jest.fn() });
		mockedUseQuery.mockReturnValue({
			data: undefined,
			isLoading: false,
			refetch: jest.fn().mockResolvedValue(undefined),
		});
		mockedUseMutation.mockReturnValue({ mutateAsync: jest.fn() });

		mockedGetNotifications.mockResolvedValue({ items: [] });
		mockedGetUnreadCount.mockResolvedValue({ unread_count: 0 });
		mockedMarkNotificationAsRead.mockResolvedValue(undefined);
		mockedMarkAllNotificationsAsRead.mockResolvedValue(undefined);

		render(
			<NotificationProvider>
				<Consumer />
			</NotificationProvider>
		);

		const queryOptions = mockedUseQuery.mock.calls.map((call) => call[0]);
		const unreadQuery = queryOptions.find((options) => options.queryKey[0] === 'notifications-unread-count');
		const listQuery = queryOptions.find((options) => options.queryKey[0] === 'notifications');
		expect(unreadQuery).toBeDefined();
		expect(listQuery).toBeDefined();

		await unreadQuery.queryFn();
		await listQuery.queryFn();

		expect(mockedGetUnreadCount).toHaveBeenCalled();
		expect(mockedGetNotifications).toHaveBeenCalledWith({ page: 1, pageSize: 5 });

		const mutationOptions = mockedUseMutation.mock.calls.map((call) => call[0]);
		await mutationOptions[0].mutationFn(11);
		await mutationOptions[1].mutationFn();

		expect(mockedMarkNotificationAsRead).toHaveBeenCalledWith(11);
		expect(mockedMarkAllNotificationsAsRead).toHaveBeenCalled();
	});
});
