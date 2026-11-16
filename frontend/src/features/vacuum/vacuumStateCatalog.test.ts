import {
    isCleaningSessionState,
    toVacuumState,
    vacuumStateLabelKeys,
} from './vacuumStateCatalog';

describe('vacuumStateCatalog', () => {
    it.each(Object.keys(vacuumStateLabelKeys))('keeps the reported state %s', (reportedState) => {
        expect(toVacuumState(reportedState)).toBe(reportedState);
    });

    it('maps absent and unrecognized states to unknown', () => {
        expect(toVacuumState(undefined)).toBe('unknown');
        expect(toVacuumState('')).toBe('unknown');
        expect(toVacuumState('warp_drive')).toBe('unknown');
    });

    it('recognizes only cleaning and paused as a cleaning session', () => {
        expect(isCleaningSessionState('cleaning')).toBe(true);
        expect(isCleaningSessionState('paused')).toBe(true);
        expect(isCleaningSessionState('returning')).toBe(false);
        expect(isCleaningSessionState('unknown')).toBe(false);
    });
});
