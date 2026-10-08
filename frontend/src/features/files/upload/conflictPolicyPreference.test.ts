import {
    conflictPolicyStorageKey,
    defaultConflictPolicy,
    loadConflictPolicy,
    saveConflictPolicy,
} from './conflictPolicyPreference';

describe('conflictPolicyPreference', () => {
    beforeEach(() => window.localStorage.clear());

    it('defaults to rename when nothing is stored', () => {
        expect(loadConflictPolicy()).toBe(defaultConflictPolicy);
        expect(defaultConflictPolicy).toBe('rename');
    });

    it('round-trips a saved policy', () => {
        saveConflictPolicy('skip');
        expect(loadConflictPolicy()).toBe('skip');
    });

    it('ignores an invalid stored value', () => {
        window.localStorage.setItem(conflictPolicyStorageKey, 'overwrite');
        expect(loadConflictPolicy()).toBe('rename');
    });

    it('survives a storage that throws', () => {
        const getSpy = jest.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('denied');
        });
        const setSpy = jest.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('denied');
        });

        expect(loadConflictPolicy()).toBe('rename');
        expect(() => saveConflictPolicy('replace')).not.toThrow();

        getSpy.mockRestore();
        setSpy.mockRestore();
    });
});
