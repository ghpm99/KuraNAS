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

describe('components/notifications/NotificationItem', () => {
	beforeEach(() => {
		jest.useFakeTimers();
		jest.setSystemTime(new Date('2026-04-01T12:00:00.000Z'));
	});

	afterEach(() => {
		jest.useRealTimers();
	});

	it('renders title/message and handles click when callback is provided', () => {
		const onClick = jest.fn();

		render(<NotificationItem notification={BASE_NOTIFICATION} onClick={onClick} />);

		expect(screen.getByText('Title')).toBeInTheDocument();
		expect(screen.getByText('Message')).toBeInTheDocument();
		expect(screen.getByText('now')).toBeInTheDocument();

		fireEvent.click(screen.getByText('Title'));
		expect(onClick).toHaveBeenCalled();
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
		expect(screen.getByText('5m')).toBeInTheDocument();

		rerender(
			<NotificationItem
				notification={{
					...BASE_NOTIFICATION,
					created_at: '2026-04-01T09:00:00.000Z',
				}}
			/>
		);
		expect(screen.getByText('3h')).toBeInTheDocument();

		rerender(
			<NotificationItem
				notification={{
					...BASE_NOTIFICATION,
					created_at: '2026-03-30T12:00:00.000Z',
				}}
			/>
		);
		expect(screen.getByText('2d')).toBeInTheDocument();

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
