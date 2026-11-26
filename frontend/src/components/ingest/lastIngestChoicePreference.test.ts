import {
    lastIngestChoiceStorageKey,
    loadLastIngestChoice,
    saveLastIngestChoice,
} from './lastIngestChoicePreference';

const choice = { targetRoot: '/data/music', subfolder: 'Rock', preset: 'audio' };

describe('lastIngestChoicePreference', () => {
    beforeEach(() => window.localStorage.clear());

    it('returns null when nothing is stored', () => {
        expect(loadLastIngestChoice()).toBeNull();
    });

    it('round-trips a saved choice', () => {
        saveLastIngestChoice(choice);
        expect(loadLastIngestChoice()).toEqual(choice);
    });

    it('returns null for invalid JSON', () => {
        window.localStorage.setItem(lastIngestChoiceStorageKey, '{not json');
        expect(loadLastIngestChoice()).toBeNull();
    });

    it('returns null for a non-object payload', () => {
        window.localStorage.setItem(lastIngestChoiceStorageKey, 'null');
        expect(loadLastIngestChoice()).toBeNull();
        window.localStorage.setItem(lastIngestChoiceStorageKey, '"text"');
        expect(loadLastIngestChoice()).toBeNull();
    });

    it('returns null when a field is missing or has the wrong type', () => {
        window.localStorage.setItem(
            lastIngestChoiceStorageKey,
            JSON.stringify({ targetRoot: '/data', preset: 'best' })
        );
        expect(loadLastIngestChoice()).toBeNull();
        window.localStorage.setItem(
            lastIngestChoiceStorageKey,
            JSON.stringify({ ...choice, preset: 3 })
        );
        expect(loadLastIngestChoice()).toBeNull();
    });

    it('survives a storage that throws', () => {
        const getSpy = jest.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('denied');
        });
        const setSpy = jest.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('denied');
        });

        expect(loadLastIngestChoice()).toBeNull();
        expect(() => saveLastIngestChoice(choice)).not.toThrow();

        getSpy.mockRestore();
        setSpy.mockRestore();
    });
});
