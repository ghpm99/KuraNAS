import { fireEvent, render, screen } from '@testing-library/react';
import NotificationItem from './NotificationItem';

const BASE_NOTIFICATION = {
	id: 1,
	type: 'info' as const,
	title: 'Title',
	message: 'Message',
	is_read: false,
	created_at: '2026-04-01T12:00:00.000Z',
	group_count: 1,
	is_grouped: false,
};

const relative = (amount: number, unit: Intl.RelativeTimeFormatUnit) =>
	new Intl.RelativeTimeFormat('en-US', { numeric: 'auto', style: 'narrow' }).format(amount, unit);

describe('components/notifications/NotificationItem', () => {
	beforeEach(() => {
		document.documentElement.lang = 'en-US';
		jest.useFakeTimers();
		jest.setSystemTime(new Date('2026-04-01T12:00:00.000Z'));
	});

	afterEach(() => {
		document.documentElement.lang = '';
		jest.useRealTimers();
	});

	it('renders without i18n provider and without any service mock', () => {
		render(<NotificationItem notification={BASE_NOTIFICATION} />);

		expect(screen.getByText('Title')).toBeInTheDocument();
	});

	it('renders title/message and handles click when callback is provided', () => {
		const onClick = jest.fn();

		render(<NotificationItem notification={BASE_NOTIFICATION} onClick={onClick} />);

		expect(screen.getByText('Title')).toBeInTheDocument();
		expect(screen.getByText('Message')).toBeInTheDocument();
		expect(screen.getByText('now')).toBeInTheDocument();

		fireEvent.click(screen.getByRole('button'));
		expect(onClick).toHaveBeenCalled();
	});

	it('is not a button when there is no click handler and flags unread items to assistive tech', () => {
		render(<NotificationItem notification={BASE_NOTIFICATION} />);

		expect(screen.queryByRole('button')).not.toBeInTheDocument();
		expect(screen.getByText('UNREAD')).toBeInTheDocument();
	});

	it('does not flag read items as unread', () => {
		render(<NotificationItem notification={{ ...BASE_NOTIFICATION, is_read: true }} />);

		expect(screen.queryByText('UNREAD')).not.toBeInTheDocument();
	});

	it('shows grouped badge only when grouped and count is greater than one', () => {
		const grouped = {
			...BASE_NOTIFICATION,
			is_grouped: true,
			group_count: 3,
		};

		const { rerender } = render(<NotificationItem notification={grouped} />);
		expect(screen.getByText('x3')).toBeInTheDocument();

		rerender(
			<NotificationItem
				notification={{
					...BASE_NOTIFICATION,
					is_grouped: true,
					group_count: 1,
				}}
			/>
		);

		expect(screen.queryByText('x1')).not.toBeInTheDocument();
	});

	it('formats relative time for minutes, hours, days and older dates', () => {
		const { rerender } = render(
			<NotificationItem
				notification={{
					...BASE_NOTIFICATION,
					created_at: '2026-04-01T11:55:00.000Z',
				}}
			/>
		);
		expect(screen.getByText(relative(-5, 'minute'))).toBeInTheDocument();

		rerender(
			<NotificationItem
				notification={{
					...BASE_NOTIFICATION,
					created_at: '2026-04-01T09:00:00.000Z',
				}}
			/>
		);
		expect(screen.getByText(relative(-3, 'hour'))).toBeInTheDocument();

		rerender(
			<NotificationItem
				notification={{
					...BASE_NOTIFICATION,
					created_at: '2026-03-30T12:00:00.000Z',
				}}
			/>
		);
		expect(screen.getByText(relative(-2, 'day'))).toBeInTheDocument();

		const oldDate = '2026-01-01T12:00:00.000Z';
		rerender(
			<NotificationItem
				notification={{
					...BASE_NOTIFICATION,
					created_at: oldDate,
				}}
			/>
		);

		expect(screen.getByText(new Date(oldDate).toLocaleDateString())).toBeInTheDocument();
	});

	it('falls back to info config when notification type is unknown at runtime', () => {
		const unknownTypeNotification = {
			...BASE_NOTIFICATION,
			type: 'unknown' as any,
		};

		render(<NotificationItem notification={unknownTypeNotification} />);
		expect(screen.getByText('Title')).toBeInTheDocument();
	});
});
