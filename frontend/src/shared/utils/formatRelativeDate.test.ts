import { formatRelativeDate } from './formatRelativeDate';

const now = new Date('2026-11-25T12:00:00Z');
const hoursBefore = (hours: number) => new Date(now.getTime() - hours * 3_600_000).toISOString();

describe('formatRelativeDate', () => {
    beforeEach(() => {
        document.documentElement.lang = 'en-US';
    });

    it('returns an empty string for absent or invalid timestamps', () => {
        expect(formatRelativeDate(undefined, now)).toBe('');
        expect(formatRelativeDate('not-a-date', now)).toBe('');
    });

    it('formats recent moments relatively', () => {
        expect(formatRelativeDate(now.toISOString(), now)).toBe('now');
        expect(formatRelativeDate(hoursBefore(0.5), now)).toBe('30 minutes ago');
        expect(formatRelativeDate(hoursBefore(3), now)).toBe('3 hours ago');
        expect(formatRelativeDate(hoursBefore(48), now)).toBe('2 days ago');
        expect(formatRelativeDate(hoursBefore(7 * 24), now)).toBe('7 days ago');
    });

    it('switches to the absolute date after seven days', () => {
        expect(formatRelativeDate(hoursBefore(8 * 24), now)).toBe('Nov 17, 2026');
    });

    it('follows the document language', () => {
        document.documentElement.lang = 'pt-BR';
        expect(formatRelativeDate(hoursBefore(3), now)).toBe('há 3 horas');
    });
});
