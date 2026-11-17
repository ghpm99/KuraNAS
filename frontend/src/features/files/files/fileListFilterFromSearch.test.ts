import { parseFileListFilter } from './fileListFilterFromSearch';

describe('parseFileListFilter', () => {
    it.each([
        ['?filter=recent', 'recent'],
        ['?filter=starred', 'starred'],
    ])('reads %p as %p', (search, expectedFilter) => {
        expect(parseFileListFilter(search)).toBe(expectedFilter);
    });

    it.each(['', '?filter=', '?filter=all', '?filter=unknown', '?other=recent'])(
        'falls back to all for %p',
        (search) => {
            expect(parseFileListFilter(search)).toBe('all');
        }
    );
});
