import {
    defaultVideoPlaybackPreferences,
    loadVideoPlaybackPreferences,
    saveVideoPlaybackPreferences,
    videoPlaybackPreferencesStorageKey,
} from './videoPlaybackPreferences';

const storePreferences = (stored: unknown) =>
    window.localStorage.setItem(videoPlaybackPreferencesStorageKey, JSON.stringify(stored));

describe('videoPlaybackPreferences', () => {
    beforeEach(() => {
        window.localStorage.clear();
        jest.restoreAllMocks();
    });

    it('returns the defaults when nothing is stored', () => {
        expect(loadVideoPlaybackPreferences()).toEqual(defaultVideoPlaybackPreferences);
    });

    it('reads valid stored preferences', () => {
        storePreferences({ volume: 0.4, isMuted: true, playbackRate: 1.5 });

        expect(loadVideoPlaybackPreferences()).toEqual({
            volume: 0.4,
            isMuted: true,
            playbackRate: 1.5,
        });
    });

    it('accepts the boundary values', () => {
        storePreferences({ volume: 0, isMuted: false, playbackRate: 0.25 });
        expect(loadVideoPlaybackPreferences().playbackRate).toBe(0.25);

        storePreferences({ volume: 1, isMuted: false, playbackRate: 2 });
        expect(loadVideoPlaybackPreferences().playbackRate).toBe(2);
    });

    it.each([
        ['volume above range', { volume: 1.5, isMuted: false, playbackRate: 1 }],
        ['negative volume', { volume: -0.1, isMuted: false, playbackRate: 1 }],
        ['volume as string', { volume: '0.5', isMuted: false, playbackRate: 1 }],
        ['rate below range', { volume: 1, isMuted: false, playbackRate: 0.1 }],
        ['rate above range', { volume: 1, isMuted: false, playbackRate: 3 }],
        ['muted as string', { volume: 1, isMuted: 'yes', playbackRate: 1 }],
        ['missing fields', { volume: 1 }],
        ['null', null],
        ['array of numbers', [1, 2]],
    ])('falls back to defaults for %s', (_label, stored) => {
        storePreferences(stored);

        expect(loadVideoPlaybackPreferences()).toEqual(defaultVideoPlaybackPreferences);
    });

    it('falls back to defaults for broken JSON', () => {
        window.localStorage.setItem(videoPlaybackPreferencesStorageKey, '{not json');

        expect(loadVideoPlaybackPreferences()).toEqual(defaultVideoPlaybackPreferences);
    });

    it('survives a storage that throws', () => {
        jest.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('denied');
        });
        jest.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('denied');
        });

        expect(loadVideoPlaybackPreferences()).toEqual(defaultVideoPlaybackPreferences);
        expect(() => saveVideoPlaybackPreferences(defaultVideoPlaybackPreferences)).not.toThrow();
    });

    it('round-trips saved preferences under a single key', () => {
        const preferences = { volume: 0.25, isMuted: false, playbackRate: 0.75 };

        saveVideoPlaybackPreferences(preferences);

        expect(window.localStorage.length).toBe(1);
        expect(loadVideoPlaybackPreferences()).toEqual(preferences);
    });
});
