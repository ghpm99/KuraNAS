import {
    getContinuousProgressSeconds,
    getPlayThresholdSeconds,
    hasReachedPlayThreshold,
} from './playThreshold';

describe('playThreshold', () => {
    it('uses 30 seconds for tracks longer than a minute', () => {
        expect(getPlayThresholdSeconds(240)).toBe(30);
        expect(getPlayThresholdSeconds(60)).toBe(30);
    });

    it('uses half of the track for tracks shorter than a minute', () => {
        expect(getPlayThresholdSeconds(40)).toBe(20);
        expect(getPlayThresholdSeconds(10)).toBe(5);
    });

    it('falls back to 30 seconds while the duration is unknown', () => {
        expect(getPlayThresholdSeconds(0)).toBe(30);
        expect(getPlayThresholdSeconds(Number.NaN)).toBe(30);
        expect(getPlayThresholdSeconds(Number.POSITIVE_INFINITY)).toBe(30);
    });

    it('reaches the threshold exactly at the limit', () => {
        expect(hasReachedPlayThreshold(29.9, 240)).toBe(false);
        expect(hasReachedPlayThreshold(30, 240)).toBe(true);
        expect(hasReachedPlayThreshold(19.9, 40)).toBe(false);
        expect(hasReachedPlayThreshold(20, 40)).toBe(true);
    });

    it('counts only small forward steps as continuous playback', () => {
        expect(getContinuousProgressSeconds(10, 10.25)).toBeCloseTo(0.25);
        expect(getContinuousProgressSeconds(10, 13)).toBe(3);
        expect(getContinuousProgressSeconds(10, 60)).toBe(0);
        expect(getContinuousProgressSeconds(60, 10)).toBe(0);
        expect(getContinuousProgressSeconds(10, 10)).toBe(0);
    });
});
