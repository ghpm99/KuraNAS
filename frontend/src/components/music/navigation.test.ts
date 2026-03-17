import { getMusicSectionFromPath, getMusicSectionMeta, musicNavigationItems } from './navigation';

describe('music navigation helpers', () => {
    it('resolves the current section from known paths', () => {
        expect(getMusicSectionFromPath('/music')).toBe('home');
        expect(getMusicSectionFromPath('/music/folders')).toBe('folders');
    });

    it('falls back to home metadata for unknown paths', () => {
        expect(getMusicSectionFromPath('/music/unknown')).toBe('home');
        expect(getMusicSectionMeta('home')).toEqual(musicNavigationItems[0]);
    });
});
