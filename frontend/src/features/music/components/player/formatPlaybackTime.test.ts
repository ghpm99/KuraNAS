import { formatPlaybackTime } from './formatPlaybackTime';

describe('formatPlaybackTime', () => {
    it('formats minutes and padded seconds', () => {
        expect(formatPlaybackTime(65)).toBe('1:05');
    });

    it('falls back for non-finite values', () => {
        expect(formatPlaybackTime(NaN)).toBe('0:00');
        expect(formatPlaybackTime(Infinity)).toBe('0:00');
    });
});
