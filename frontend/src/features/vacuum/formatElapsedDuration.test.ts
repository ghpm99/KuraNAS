import { formatElapsedDuration } from './formatElapsedDuration';

describe('formatElapsedDuration', () => {
    it.each([
        [0, '0:00'],
        [5, '0:05'],
        [754, '12:34'],
        [3599, '59:59'],
        [3600, '1:00:00'],
        [3725, '1:02:05'],
        [36000, '10:00:00'],
    ])('formats %s seconds as %s', (seconds, expected) => {
        expect(formatElapsedDuration(seconds)).toBe(expected);
    });

    it('drops the fraction of a second', () => {
        expect(formatElapsedDuration(61.9)).toBe('1:01');
    });

    it.each([undefined, Number.NaN, -30])('treats %s as zero', (seconds) => {
        expect(formatElapsedDuration(seconds)).toBe('0:00');
    });
});
