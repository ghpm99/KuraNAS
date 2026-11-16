import { createNetworkFailure, createServerFailure } from './testSupport/createServerFailure';
import { isVacuumNotConfiguredError, resolveVacuumFailureMessage } from './vacuumFailures';

describe('vacuumFailures', () => {
    it('recognizes the 409 answer as not configured', () => {
        expect(isVacuumNotConfiguredError(createServerFailure(409, 'x'))).toBe(true);
    });

    it('does not take other failures as not configured', () => {
        expect(isVacuumNotConfiguredError(createServerFailure(502, 'x'))).toBe(false);
        expect(isVacuumNotConfiguredError(createNetworkFailure())).toBe(false);
        expect(isVacuumNotConfiguredError(null)).toBe(false);
    });

    it('uses the server message as it came', () => {
        expect(resolveVacuumFailureMessage(createServerFailure(502, 'robot offline'), 'offline')).toBe(
            'robot offline'
        );
    });

    it('falls back to the connection message without a server answer', () => {
        expect(resolveVacuumFailureMessage(createNetworkFailure(), 'no connection')).toBe(
            'no connection'
        );
        expect(resolveVacuumFailureMessage(undefined, 'no connection')).toBe('no connection');
    });
});
