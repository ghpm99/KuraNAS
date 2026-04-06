jest.mock('./index', () => ({
	apiBase: {
		get: jest.fn(),
		put: jest.fn(),
	},
}));

import { apiBase } from './index';
import {
	getNotifications,
	getUnreadCount,
	markNotificationAsRead,
	markAllNotificationsAsRead,
} from './notifications';

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
	put: jest.Mock;
};

describe('service/notifications', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	describe('getNotifications', () => {
		it('fetches notifications with default params', async () => {
			const data = { items: [], total: 0 };
			mockedApi.get.mockResolvedValue({ data });
			const result = await getNotifications();
			expect(mockedApi.get).toHaveBeenCalledWith('/notifications', {
				params: {
					page: 1,
					page_size: 20,
					type: undefined,
					is_read: undefined,
				},
			});
			expect(result).toEqual(data);
		});

		it('fetches notifications with custom params', async () => {
			const data = { items: [], total: 0 };
			mockedApi.get.mockResolvedValue({ data });
			const result = await getNotifications({
				page: 2,
				pageSize: 10,
				type: 'info',
				is_read: true,
			});
			expect(mockedApi.get).toHaveBeenCalledWith('/notifications', {
				params: {
					page: 2,
					page_size: 10,
					type: 'info',
					is_read: true,
				},
			});
			expect(result).toEqual(data);
		});
	});

	it('fetches unread count', async () => {
		const data = { count: 5 };
		mockedApi.get.mockResolvedValue({ data });
		const result = await getUnreadCount();
		expect(mockedApi.get).toHaveBeenCalledWith('/notifications/unread-count');
		expect(result).toEqual(data);
	});

	it('marks a notification as read', async () => {
		mockedApi.put.mockResolvedValue({});
		await markNotificationAsRead(42);
		expect(mockedApi.put).toHaveBeenCalledWith('/notifications/42/read');
	});

	it('marks all notifications as read', async () => {
		mockedApi.put.mockResolvedValue({});
		await markAllNotificationsAsRead();
		expect(mockedApi.put).toHaveBeenCalledWith('/notifications/read-all');
	});
});
