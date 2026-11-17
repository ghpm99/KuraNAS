import { parseFileListFilter } from './fileListFilterFromSearch';

describe('parseFileListFilter', () => {
    it('reads the recent filter from the query string', () => {
        expect(parseFileListFilter('?filter=recent')).toBe('recent');
    });

    it.each(['', '?filter=starred', '?filter=', '?other=recent'])(
        'falls back to all for %p',
        (search) => {
            expect(parseFileListFilter(search)).toBe('all');
        }
    );
});
