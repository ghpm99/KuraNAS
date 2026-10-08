import { readOptionalDate } from './optionalDate';

describe('readOptionalDate', () => {
    it('reads the backend optional wrapper', () => {
        expect(readOptionalDate({ Value: '2026-02-01T10:00:00Z', HasValue: true })).toBe(
            '2026-02-01T10:00:00Z'
        );
    });

    it('returns null when the wrapper has no value', () => {
        expect(readOptionalDate({ Value: '0001-01-01T00:00:00Z', HasValue: false })).toBeNull();
        expect(readOptionalDate({ Value: 12, HasValue: true })).toBeNull();
    });

    it('accepts plain date strings and rejects empty or zero dates', () => {
        expect(readOptionalDate('2026-02-01T10:00:00Z')).toBe('2026-02-01T10:00:00Z');
        expect(readOptionalDate('')).toBeNull();
        expect(readOptionalDate('0001-01-01T00:00:00Z')).toBeNull();
    });

    it('tolerates garbage input', () => {
        expect(readOptionalDate(undefined)).toBeNull();
        expect(readOptionalDate(null)).toBeNull();
        expect(readOptionalDate(42)).toBeNull();
        expect(readOptionalDate({})).toBeNull();
    });
});
