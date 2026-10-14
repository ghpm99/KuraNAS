import {
    autoplayNextPreferenceStorageKey,
    loadAutoplayNextPreference,
    saveAutoplayNextPreference,
} from './autoplayNextPreference';

describe('autoplayNextPreference', () => {
    beforeEach(() => {
        window.localStorage.clear();
        jest.restoreAllMocks();
    });

    it('returns the fallback when nothing is stored', () => {
        expect(loadAutoplayNextPreference(true)).toBe(true);
        expect(loadAutoplayNextPreference(false)).toBe(false);
    });

    it('persists and reads back the stored value over the fallback', () => {
        saveAutoplayNextPreference(false);

        expect(loadAutoplayNextPreference(true)).toBe(false);
    });

    it('ignores stored values that are not booleans', () => {
        window.localStorage.setItem(autoplayNextPreferenceStorageKey, '"yes"');
        expect(loadAutoplayNextPreference(true)).toBe(true);

        window.localStorage.setItem(autoplayNextPreferenceStorageKey, '{broken');
        expect(loadAutoplayNextPreference(false)).toBe(false);
    });

    it('survives storage failures', () => {
        jest.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('denied');
        });
        jest.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('denied');
        });

        expect(loadAutoplayNextPreference(true)).toBe(true);
        expect(() => saveAutoplayNextPreference(false)).not.toThrow();
    });
});
