import { parseDuplicatesType } from './duplicatesTypeFilter';

describe('parseDuplicatesType', () => {
    it('reads the image filter from the query string', () => {
        expect(parseDuplicatesType('?type=image')).toBe('image');
    });

    it('ignores absent or unknown types', () => {
        expect(parseDuplicatesType('')).toBeUndefined();
        expect(parseDuplicatesType('?type=video')).toBeUndefined();
        expect(parseDuplicatesType('?period=7d')).toBeUndefined();
    });
});
