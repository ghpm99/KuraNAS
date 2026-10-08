import { formatNotificationRelativeTime } from './notificationRelativeTime';

const now = new Date('2026-04-01T12:00:00.000Z');

describe('components/notifications/notificationRelativeTime', () => {
	afterEach(() => {
		document.documentElement.lang = '';
	});

	it('uses the document language for relative units', () => {
		document.documentElement.lang = 'pt-BR';
		const expectedMinutes = new Intl.RelativeTimeFormat('pt-BR', {
			numeric: 'auto',
			style: 'narrow',
		}).format(-5, 'minute');

		expect(formatNotificationRelativeTime('2026-04-01T11:55:00.000Z', now)).toBe(expectedMinutes);
	});

	it('falls back to the default locale when the document language is invalid', () => {
		document.documentElement.lang = 'not a locale!!';

		expect(formatNotificationRelativeTime('2026-04-01T11:59:30.000Z', now)).not.toBe('');
	});

	it('returns an empty string for an invalid date', () => {
		expect(formatNotificationRelativeTime('not-a-date', now)).toBe('');
	});

	it('clamps future timestamps to now and prints a date after thirty days', () => {
		document.documentElement.lang = 'en-US';
		const justNow = new Intl.RelativeTimeFormat('en-US', {
			numeric: 'auto',
			style: 'narrow',
		}).format(0, 'second');

		expect(formatNotificationRelativeTime('2026-04-02T12:00:00.000Z', now)).toBe(justNow);
		expect(formatNotificationRelativeTime('2026-01-01T12:00:00.000Z', now)).toBe(
			new Date('2026-01-01T12:00:00.000Z').toLocaleDateString('en-US')
		);
	});
});
