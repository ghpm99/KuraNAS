import { getVideoClassificationLabelKey } from './videoClassificationLabel';

describe('getVideoClassificationLabelKey', () => {
    it.each([
        ['series', 'VIDEO_CLASS_SERIES'],
        ['movie', 'VIDEO_CLASS_MOVIE'],
        ['anime', 'VIDEO_CLASS_ANIME'],
        ['course', 'VIDEO_CLASS_COURSE'],
        ['personal', 'VIDEO_CLASS_PERSONAL'],
        ['clip', 'VIDEO_CLASS_CLIP'],
        ['music_video', 'VIDEO_CLASS_MUSIC_VIDEO'],
        ['program', 'VIDEO_CLASS_PROGRAM'],
        ['SERIES', 'VIDEO_CLASS_SERIES'],
    ])('maps %s to %s', (classification, expectedKey) => {
        expect(getVideoClassificationLabelKey(classification)).toBe(expectedKey);
    });

    it.each([['unknown_kind'], [''], [undefined], [null]])(
        'falls back to the generic videos key for %s',
        (classification) => {
            expect(getVideoClassificationLabelKey(classification)).toBe('NAV_VIDEOS');
        }
    );
});
