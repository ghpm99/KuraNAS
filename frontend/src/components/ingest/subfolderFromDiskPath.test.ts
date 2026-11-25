import { subfolderFromDiskPath } from './subfolderFromDiskPath';

describe('subfolderFromDiskPath', () => {
    it('returns the relative path inside the root', () => {
        expect(subfolderFromDiskPath('/data/videos/', '/data/videos/shows/a')).toBe('shows/a');
    });

    it('returns empty for the root itself', () => {
        expect(subfolderFromDiskPath('/data/videos', '/data/videos/')).toBe('');
    });

    it('handles windows separators', () => {
        expect(subfolderFromDiskPath('D:\\media', 'D:\\media\\clips')).toBe('clips');
    });

    it('returns null outside the root, including sibling prefixes', () => {
        expect(subfolderFromDiskPath('/data/videos', '/data/videos2/x')).toBeNull();
        expect(subfolderFromDiskPath('/data/videos', '/other')).toBeNull();
    });
});
