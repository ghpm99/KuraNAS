import { computeNextShutdownExecution } from './nextShutdownExecution';

describe('settings/nextShutdownExecution', () => {
    const now = new Date(2026, 10, 17, 10, 0, 0);

    it('returns today when the time is still ahead', () => {
        expect(computeNextShutdownExecution('23:30', now)).toEqual(
            new Date(2026, 10, 17, 23, 30, 0)
        );
    });

    it('returns tomorrow when the time has passed', () => {
        expect(computeNextShutdownExecution('03:00', now)).toEqual(new Date(2026, 10, 18, 3, 0, 0));
    });

    it('returns null for malformed or out of range times', () => {
        expect(computeNextShutdownExecution('', now)).toBeNull();
        expect(computeNextShutdownExecution('abc', now)).toBeNull();
        expect(computeNextShutdownExecution('25:00', now)).toBeNull();
    });
});
