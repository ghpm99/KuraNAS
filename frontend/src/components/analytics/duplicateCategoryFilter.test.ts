import { parseDuplicateCategory } from './duplicateCategoryFilter';

describe('parseDuplicateCategory', () => {
    it.each(['image', 'audio', 'video', 'document', 'other'])('reads %s', (category) => {
        expect(parseDuplicateCategory(`?type=${category}`)).toBe(category);
    });

    it.each(['', '?type=', '?type=archive', '?period=7d'])('ignores %p', (search) => {
        expect(parseDuplicateCategory(search)).toBeUndefined();
    });
});
