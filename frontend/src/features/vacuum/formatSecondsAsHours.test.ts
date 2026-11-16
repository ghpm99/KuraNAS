import { formatSecondsAsHours } from './formatSecondsAsHours';

describe('formatSecondsAsHours', () => {
    it.each([
        [0, '0.0'],
        [1800, '0.5'],
        [3600, '1.0'],
        [803013, '223.1'],
        [1080000, '300.0'],
    ])('formats %s seconds as %s hours', (seconds, expected) => {
        expect(formatSecondsAsHours(seconds)).toBe(expected);
    });

    it.each([undefined, Number.NaN, -3600])('treats %s as zero', (seconds) => {
        expect(formatSecondsAsHours(seconds)).toBe('0.0');
    });
});
