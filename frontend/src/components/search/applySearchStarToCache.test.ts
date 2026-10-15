import type { GlobalSearchResponse } from '@/service/search';
import { applySearchStarToCache } from './applySearchStarToCache';

const buildResponse = (): GlobalSearchResponse =>
    ({
        query: 'q',
        files: [
            { id: 1, starred: false },
            { id: 2, starred: false },
        ],
        folders: [{ id: 3, starred: true }],
        artists: [],
        albums: [],
        playlists: [],
        videos: [],
        images: [],
    }) as unknown as GlobalSearchResponse;

describe('applySearchStarToCache', () => {
    it('returns undefined when nothing is cached', () => {
        expect(applySearchStarToCache(undefined, 1, true)).toBeUndefined();
    });

    it('updates only the matching file', () => {
        const updated = applySearchStarToCache(buildResponse(), 1, true);
        expect(updated?.files.map((file) => file.starred)).toEqual([true, false]);
    });

    it('updates the matching folder', () => {
        const updated = applySearchStarToCache(buildResponse(), 3, false);
        expect(updated?.folders[0]?.starred).toBe(false);
    });

    it('tolerates a partial payload without files or folders', () => {
        const partial = { query: 'q' } as unknown as GlobalSearchResponse;
        expect(applySearchStarToCache(partial, 1, true)).toEqual(partial);
    });
});
