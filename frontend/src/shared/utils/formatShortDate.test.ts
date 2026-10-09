import { formatShortDate } from './formatShortDate';

describe('formatShortDate', () => {
    it('formats a valid ISO timestamp as a medium date', () => {
        expect(formatShortDate('2026-03-04T12:00:00Z', 'en-US')).toBe('Mar 4, 2026');
    });

    it('returns an empty string for an invalid or missing timestamp', () => {
        expect(formatShortDate('not-a-date', 'en-US')).toBe('');
        expect(formatShortDate('', 'en-US')).toBe('');
    });
});
