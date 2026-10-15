import { applyVideoStarToCache } from './applyVideoStarToCache';

const buildVideo = (id: number, starred = false) => ({
    id,
    name: `video-${id}.mkv`,
    path: `/videos/video-${id}.mkv`,
    format: 'mkv',
    starred,
});

describe('applyVideoStarToCache', () => {
    it('returns primitives and nullish values untouched', () => {
        expect(applyVideoStarToCache(undefined, 1, true)).toBeUndefined();
        expect(applyVideoStarToCache(null, 1, true)).toBeNull();
        expect(applyVideoStarToCache('text', 1, true)).toBe('text');
    });

    it('stars the matching video nested in pages, playlist items and arrays', () => {
        const cache = {
            pages: [{ items: [buildVideo(1), buildVideo(2)] }],
            items: [{ id: 99, video: buildVideo(1) }],
        };

        const updated = applyVideoStarToCache(cache, 1, true);

        expect(updated.pages[0]!.items.map((video) => video.starred)).toEqual([true, false]);
        expect(updated.items[0]!.video.starred).toBe(true);
        expect(cache.pages[0]!.items[0]!.starred).toBe(false);
    });

    it('does not touch records that only share the id but are not videos', () => {
        const playlist = { id: 1, name: 'Playlist', source_path: '/x' };

        expect(applyVideoStarToCache({ playlist }, 1, true)).toEqual({ playlist });
    });
});
