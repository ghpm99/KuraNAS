import {
    conflictPolicyStorageKey,
    defaultConflictPolicy,
    loadConflictPolicy,
    saveConflictPolicy,
    toWireConflictPolicy,
} from './conflictPolicyPreference';

describe('conflictPolicyPreference', () => {
    beforeEach(() => window.localStorage.clear());

    it('defaults to ask when nothing is stored', () => {
        expect(loadConflictPolicy()).toBe(defaultConflictPolicy);
        expect(defaultConflictPolicy).toBe('ask');
    });

    it('round-trips a saved policy', () => {
        saveConflictPolicy('skip');
        expect(loadConflictPolicy()).toBe('skip');
    });

    it('ignores an invalid stored value', () => {
        window.localStorage.setItem(conflictPolicyStorageKey, 'overwrite');
        expect(loadConflictPolicy()).toBe('ask');
    });

    it('survives a storage that throws', () => {
        const getSpy = jest.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('denied');
        });
        const setSpy = jest.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('denied');
        });

        expect(loadConflictPolicy()).toBe('ask');
        expect(() => saveConflictPolicy('replace')).not.toThrow();

        getSpy.mockRestore();
        setSpy.mockRestore();
    });

    it('round-trips every choice including ask', () => {
        (['ask', 'rename', 'replace', 'skip'] as const).forEach((choice) => {
            saveConflictPolicy(choice);
            expect(loadConflictPolicy()).toBe(choice);
        });
    });

    it('maps ask to fail on the wire and keeps the other choices', () => {
        expect(toWireConflictPolicy('ask')).toBe('fail');
        expect(toWireConflictPolicy('rename')).toBe('rename');
        expect(toWireConflictPolicy('replace')).toBe('replace');
        expect(toWireConflictPolicy('skip')).toBe('skip');
    });
});
