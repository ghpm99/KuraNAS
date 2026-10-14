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

    it('round-trips an enabled preference', () => {
        saveVideoSubtitlePreferences({ isEnabled: true, language: 'pt-BR' });

        expect(loadVideoSubtitlePreferences()).toEqual({ isEnabled: true, language: 'pt-BR' });
    });

    it('round-trips a disabled preference keeping the last language', () => {
        saveVideoSubtitlePreferences({ isEnabled: false, language: 'en' });

        expect(loadVideoSubtitlePreferences()).toEqual({ isEnabled: false, language: 'en' });
    });

    it('accepts a disabled preference without any language', () => {
        saveVideoSubtitlePreferences({ isEnabled: false, language: '' });

        expect(loadVideoSubtitlePreferences()).toEqual({ isEnabled: false, language: '' });
    });

    it('reads the legacy language-only format as enabled', () => {
        window.localStorage.setItem(videoSubtitlePreferencesStorageKey, '{"language":"pt-BR"}');

        expect(loadVideoSubtitlePreferences()).toEqual({ isEnabled: true, language: 'pt-BR' });
    });

    it.each([
        ['malformed json', '{oops'],
        ['a non-object', '42'],
        ['null', 'null'],
        ['a missing language', '{}'],
        ['a blank language', '{"language":"  "}'],
        ['a non-string language', '{"language":3}'],
        ['a non-boolean isEnabled', '{"isEnabled":"yes","language":"en"}'],
        ['an enabled preference without language', '{"isEnabled":true,"language":""}'],
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
        expect(() =>
            saveVideoSubtitlePreferences({ isEnabled: true, language: 'en' })
        ).not.toThrow();
    });
});
