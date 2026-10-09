import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import NotificationDropdown from './NotificationDropdown';

const mockNavigate = jest.fn();
const mockMarkAsRead = jest.fn();
const mockMarkAllAsRead = jest.fn();

jest.mock('react-router-dom', () => ({
	...jest.requireActual('react-router-dom'),
	useNavigate: () => mockNavigate,
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string) => key,
	}),
}));

const mockUseNotifications = jest.fn();

jest.mock('@/components/providers/notificationProvider/notificationContext', () => ({
	useNotifications: () => mockUseNotifications(),
}));

const notificationsFixture = [
	{
		id: 10,
		type: 'info' as const,
		title: 'Unread title',
		message: 'Unread message',
		is_read: false,
		created_at: '2026-04-01T12:00:00.000Z',
		group_count: 1,
		is_grouped: false,
	},
	{
		id: 20,
		type: 'success' as const,
		title: 'Read title',
		message: 'Read message',
		is_read: true,
		created_at: '2026-04-01T11:00:00.000Z',
		group_count: 1,
		is_grouped: false,
	},
];

describe('components/notifications/NotificationDropdown', () => {
	beforeEach(() => {
		mockNavigate.mockReset();
		mockMarkAsRead.mockReset();
		mockMarkAllAsRead.mockReset();
		mockUseNotifications.mockReset();
		mockUseNotifications.mockReturnValue({
			notifications: notificationsFixture,
			unreadCount: 1,
			markAsRead: mockMarkAsRead,
			markAllAsRead: mockMarkAllAsRead,
		});
	});

	it('renders notifications, marks unread item and keeps read item untouched', async () => {
		const onClose = jest.fn();
		render(<NotificationDropdown onClose={onClose} />);

		expect(screen.getByText('Unread title')).toBeInTheDocument();
		expect(screen.getByText('Read title')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'MARK_ALL_AS_READ' })).toBeInTheDocument();

		fireEvent.click(screen.getByRole('button', { name: /Unread title/ }));
		await waitFor(() => {
			expect(mockMarkAsRead).toHaveBeenCalledWith(10);
		});

		fireEvent.click(screen.getByRole('button', { name: /Read title/ }));
		expect(mockMarkAsRead).toHaveBeenCalledTimes(1);
		expect(mockNavigate).not.toHaveBeenCalled();
		expect(onClose).not.toHaveBeenCalled();
	});

	it('hides mark-all button and shows empty state when there are no notifications', () => {
		mockUseNotifications.mockReturnValue({
			notifications: [],
			unreadCount: 0,
			markAsRead: mockMarkAsRead,
			markAllAsRead: mockMarkAllAsRead,
		});

		render(<NotificationDropdown onClose={jest.fn()} />);

		expect(screen.getByText('NO_NOTIFICATIONS')).toBeInTheDocument();
		expect(screen.queryByRole('button', { name: 'MARK_ALL_AS_READ' })).not.toBeInTheDocument();
	});

	it('closes dropdown and navigates to notifications page when clicking view-all', () => {
		const onClose = jest.fn();
		render(<NotificationDropdown onClose={onClose} />);

		fireEvent.click(screen.getByRole('button', { name: 'VIEW_ALL_NOTIFICATIONS' }));

		expect(onClose).toHaveBeenCalled();
		expect(mockNavigate).toHaveBeenCalledWith('/notifications');
	});

	it('marks as read, closes and navigates to the target route when the notification carries one', async () => {
		const onClose = jest.fn();
		mockUseNotifications.mockReturnValue({
			notifications: [
				{
					...notificationsFixture[0],
					metadata: { event: 'capture_upload_completed' },
				},
			],
			unreadCount: 1,
			markAsRead: mockMarkAsRead,
			markAllAsRead: mockMarkAllAsRead,
		});
		render(<NotificationDropdown onClose={onClose} />);

		fireEvent.click(screen.getByRole('button', { name: /Unread title/ }));

		await waitFor(() => expect(mockMarkAsRead).toHaveBeenCalledWith(10));
		expect(onClose).toHaveBeenCalled();
		expect(mockNavigate).toHaveBeenCalledWith('/captures');
	});

	it('navigates without marking again when a read notification has a target route', () => {
		mockUseNotifications.mockReturnValue({
			notifications: [{ ...notificationsFixture[1], group_key: 'takeout_import' }],
			unreadCount: 0,
			markAsRead: mockMarkAsRead,
			markAllAsRead: mockMarkAllAsRead,
		});
		render(<NotificationDropdown onClose={jest.fn()} />);

		fireEvent.click(screen.getByRole('button', { name: /Read title/ }));

		expect(mockMarkAsRead).not.toHaveBeenCalled();
		expect(mockNavigate).toHaveBeenCalledWith('/takeout');
	});

	it('limits the popover width to the viewport', () => {
		const { container } = render(<NotificationDropdown onClose={jest.fn()} />);

		const rootClassName = (container.firstChild as HTMLElement).className.split(' ').find((name) => name.startsWith('css-'));
		const injectedRules = Array.from(document.styleSheets).flatMap((sheet) =>
			Array.from(sheet.cssRules).map((rule) => rule.cssText)
		);
		const rootRule = injectedRules.find((rule) => rule.includes(`.${rootClassName}`)) ?? '';
		expect(rootRule).toContain('min(360px, calc(100vw - 32px))');
	});
});
