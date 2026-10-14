import { formatVideoDuration } from './formatVideoDuration';

describe('formatVideoDuration', () => {
    it.each([
        [59, '0:59'],
        [60, '1:00'],
        [61.9, '1:01'],
        [3599, '59:59'],
        [3600, '1:00:00'],
        [3725, '1:02:05'],
        [36000, '10:00:00'],
    ])('formats %d seconds as %s', (seconds, expectedLabel) => {
        expect(formatVideoDuration(seconds)).toBe(expectedLabel);
    });

    it.each([[undefined], [null], [0], [-5], [NaN], [Infinity]])(
        'returns null for %p',
        (seconds) => {
            expect(formatVideoDuration(seconds as number | null | undefined)).toBeNull();
        }
    );
});
