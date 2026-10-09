import { getMusicSectionFromPath, getMusicSectionMeta } from './navigation';

describe('music navigation helpers', () => {
    it('resolves the current section from known paths', () => {
        expect(getMusicSectionFromPath('/music')).toBe('home');
        expect(getMusicSectionFromPath('/music/folders')).toBe('folders');
        expect(getMusicSectionFromPath('/music/tracks')).toBe('tracks');
    });

    it('resolves the search view to its own header metadata', () => {
        expect(getMusicSectionFromPath('/music/search')).toBe('search');
        expect(getMusicSectionMeta('search')).toEqual(
            expect.objectContaining({
                labelKey: 'GLOBAL_SEARCH_ACTION_MUSIC_SEARCH',
                descriptionKey: 'MUSIC_SEARCH_DESCRIPTION',
            })
        );
    });

    it('falls back to home metadata for unknown paths', () => {
        expect(getMusicSectionFromPath('/music/unknown')).toBe('home');
        expect(getMusicSectionMeta('home')).toEqual(
            expect.objectContaining({ key: 'home' })
        );
    });
});
