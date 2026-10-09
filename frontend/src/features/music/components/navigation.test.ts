import { getMusicSectionFromPath, getMusicSectionMeta } from './navigation';

describe('music navigation helpers', () => {
    it('resolves the current section from known paths', () => {
        expect(getMusicSectionFromPath('/music')).toBe('home');
        expect(getMusicSectionFromPath('/music/folders')).toBe('folders');
        expect(getMusicSectionFromPath('/music/tracks')).toBe('tracks');
    });

    it('falls back to home metadata for unknown paths', () => {
        expect(getMusicSectionFromPath('/music/unknown')).toBe('home');
        expect(getMusicSectionMeta('home')).toEqual(
            expect.objectContaining({ key: 'home' })
        );
    });
});
