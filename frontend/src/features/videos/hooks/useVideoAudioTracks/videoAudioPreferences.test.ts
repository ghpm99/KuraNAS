import {
    loadVideoAudioPreferences,
    saveVideoAudioPreferences,
    videoAudioPreferencesStorageKey,
} from './videoAudioPreferences';

describe('useVideoAudioTracks/videoAudioPreferences', () => {
    beforeEach(() => window.localStorage.clear());

    it('returns null when nothing was stored', () => {
        expect(loadVideoAudioPreferences()).toBeNull();
    });

    it('round-trips the primary language', () => {
        saveVideoAudioPreferences({ language: 'pt-BR' });

        expect(loadVideoAudioPreferences()).toEqual({ language: 'pt' });
    });

    it.each(['not json', '"text"', 'null', '{"language":3}', '{"language":""}'])(
        'ignores the malformed payload %s',
        (payload) => {
            window.localStorage.setItem(videoAudioPreferencesStorageKey, payload);

            expect(loadVideoAudioPreferences()).toBeNull();
        }
    );

    it('survives a storage that throws', () => {
        const getItem = jest.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('denied');
        });
        const setItem = jest.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('denied');
        });

        expect(loadVideoAudioPreferences()).toBeNull();
        expect(() => saveVideoAudioPreferences({ language: 'ja' })).not.toThrow();

        getItem.mockRestore();
        setItem.mockRestore();
    });
});
