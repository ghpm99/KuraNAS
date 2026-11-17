import { createTransferRateTracker } from './transferRateTracker';

const buildClock = () => {
    let currentTime = 0;
    return {
        now: () => currentTime,
        advance: (milliseconds: number) => {
            currentTime += milliseconds;
        },
    };
};

describe('createTransferRateTracker', () => {
    it('gives no estimate before one second of samples', () => {
        const clock = buildClock();
        const tracker = createTransferRateTracker(clock.now);

        expect(tracker.record(0, 1000)).toEqual({});
        clock.advance(999);
        expect(tracker.record(500, 1000)).toEqual({});
    });

    it('estimates speed and remaining time once one second of samples exists', () => {
        const clock = buildClock();
        const tracker = createTransferRateTracker(clock.now);

        tracker.record(0, 10_000);
        clock.advance(2000);

        expect(tracker.record(2000, 10_000)).toEqual({ bytesPerSecond: 1000, remainingSeconds: 8 });
    });

    it('averages only over the last five seconds', () => {
        const clock = buildClock();
        const tracker = createTransferRateTracker(clock.now);

        tracker.record(0, 100_000);
        clock.advance(5000);
        tracker.record(50_000, 100_000);
        clock.advance(5000);
        const estimate = tracker.record(51_000, 100_000);

        expect(estimate.bytesPerSecond).toBeCloseTo(200);
        expect(estimate.remainingSeconds).toBeCloseTo(245);
    });

    it('restarts the measure when the sent bytes go backwards after a realignment', () => {
        const clock = buildClock();
        const tracker = createTransferRateTracker(clock.now);

        tracker.record(0, 10_000);
        clock.advance(2000);
        tracker.record(4000, 10_000);
        clock.advance(500);
        expect(tracker.record(1000, 10_000)).toEqual({});

        clock.advance(1000);
        expect(tracker.record(2000, 10_000)).toEqual({
            bytesPerSecond: 1000,
            remainingSeconds: 8,
        });
    });

    it('gives no estimate while no bytes advanced', () => {
        const clock = buildClock();
        const tracker = createTransferRateTracker(clock.now);

        tracker.record(100, 1000);
        clock.advance(3000);

        expect(tracker.record(100, 1000)).toEqual({});
    });

    it('forgets all samples on reset', () => {
        const clock = buildClock();
        const tracker = createTransferRateTracker(clock.now);

        tracker.record(0, 1000);
        clock.advance(2000);
        tracker.reset();

        expect(tracker.record(900, 1000)).toEqual({});
    });

    it('never reports negative remaining time when sent bytes exceed the total', () => {
        const clock = buildClock();
        const tracker = createTransferRateTracker(clock.now);

        tracker.record(0, 1000);
        clock.advance(1000);

        expect(tracker.record(1200, 1000).remainingSeconds).toBe(0);
    });
});
