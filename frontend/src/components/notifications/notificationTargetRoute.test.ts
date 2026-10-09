import type { Notification } from '@/types/notification';
import { resolveNotificationTargetRoute } from './notificationTargetRoute';

const buildNotification = (overrides: Partial<Notification>): Notification => ({
	id: 1,
	type: 'info',
	title: 'Title',
	message: 'Message',
	is_read: false,
	created_at: '2026-04-01T12:00:00.000Z',
	group_count: 1,
	is_grouped: false,
	...overrides,
});

describe('components/notifications/notificationTargetRoute', () => {
	it.each([
		[{ metadata: { event: 'capture_promoted' } }, '/captures'],
		[{ metadata: { event: 'takeout_import_started' } }, '/takeout'],
		[{ metadata: { event: 'watch_folder_import' } }, '/files'],
		[{ group_key: 'capture_upload_result' }, '/captures'],
		[{ group_key: 'takeout_import' }, '/takeout'],
		[{ group_key: 'watch_import_4' }, '/files'],
		[{ group_key: 'ytdlp-update-2026.1' }, '/settings'],
	])('maps %p to %s', (overrides, expectedRoute) => {
		expect(resolveNotificationTargetRoute(buildNotification(overrides))).toBe(expectedRoute);
	});

	it.each([
		[{}],
		[{ metadata: { event: 'unknown_event' } }],
		[{ metadata: { event: 42 } }],
		[{ group_key: 'auto_shutdown' }],
	])('returns undefined for %p', (overrides) => {
		expect(resolveNotificationTargetRoute(buildNotification(overrides))).toBeUndefined();
	});

	it('falls back to the group key when the event is not routable', () => {
		const notification = buildNotification({
			metadata: { event: 'something_else' },
			group_key: 'takeout_import',
		});

		expect(resolveNotificationTargetRoute(notification)).toBe('/takeout');
	});
});
