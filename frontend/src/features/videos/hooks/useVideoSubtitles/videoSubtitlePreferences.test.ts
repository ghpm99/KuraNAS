import {
    loadVideoSubtitlePreferences,
    saveVideoSubtitlePreferences,
    videoSubtitlePreferencesStorageKey,
} from './videoSubtitlePreferences';

describe('videoSubtitlePreferences', () => {
    beforeEach(() => {
        window.localStorage.clear();
        jest.restoreAllMocks();
    });

    it('returns null when nothing is stored', () => {
        expect(loadVideoSubtitlePreferences()).toBeNull();
    });

    it('round-trips the saved language', () => {
        saveVideoSubtitlePreferences({ language: 'pt-BR' });

        expect(loadVideoSubtitlePreferences()).toEqual({ language: 'pt-BR' });
    });

    it.each([
        ['malformed json', '{oops'],
        ['a non-object', '42'],
        ['null', 'null'],
        ['a missing language', '{}'],
        ['a blank language', '{"language":"  "}'],
        ['a non-string language', '{"language":3}'],
    ])('ignores %s', (_name, storedValue) => {
        window.localStorage.setItem(videoSubtitlePreferencesStorageKey, storedValue);

        expect(loadVideoSubtitlePreferences()).toBeNull();
    });

    it('survives storage that throws', () => {
        jest.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('blocked');
        });
        jest.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('blocked');
        });

        expect(loadVideoSubtitlePreferences()).toBeNull();
        expect(() => saveVideoSubtitlePreferences({ language: 'en' })).not.toThrow();
    });
});
