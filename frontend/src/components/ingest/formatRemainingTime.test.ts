import { formatRemainingTime } from './formatRemainingTime';

describe('formatRemainingTime', () => {
    it('formats seconds, minutes and hours', () => {
        expect(formatRemainingTime(42)).toBe('42s');
        expect(formatRemainingTime(125)).toBe('2m 5s');
        expect(formatRemainingTime(3725)).toBe('1h 2m 5s');
    });
});
