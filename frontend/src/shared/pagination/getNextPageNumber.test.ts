import { getNextPageNumber } from './getNextPageNumber';

const buildPage = (page: number, hasNext: boolean) => ({
    items: [],
    pagination: { page, page_size: 10, has_next: hasNext, has_prev: page > 1 },
});

describe('getNextPageNumber', () => {
    it('returns the following page while the server reports has_next', () => {
        expect(getNextPageNumber(buildPage(3, true))).toBe(4);
    });

    it('stops on the last page', () => {
        expect(getNextPageNumber(buildPage(3, false))).toBeUndefined();
    });

    it('survives an absent or partial response', () => {
        expect(getNextPageNumber(undefined)).toBeUndefined();
        expect(getNextPageNumber({} as never)).toBeUndefined();
    });
});
