import { formatDateTime } from './formatDateTime';

describe('formatDateTime', () => {
    it('writes the day before the month in pt-BR', () => {
        expect(formatDateTime('2026-06-11T10:00:00', 'pt-BR')).toMatch(/^11\/06\/2026/);
    });

    it('writes the month before the day in en-US', () => {
        expect(formatDateTime('2026-06-11T10:00:00', 'en-US')).toMatch(/^6\/11\/26/);
    });

    it('is empty for a missing or unreadable timestamp', () => {
        expect(formatDateTime(undefined, 'pt-BR')).toBe('');
        expect(formatDateTime('not a date', 'pt-BR')).toBe('');
    });
});
