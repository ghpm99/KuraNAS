import {
    countActiveFileSearchFilters,
    emptyFileSearchFilters,
    parseFileSearchFilters,
    toFileSearchRefinements,
    writeFileSearchFilters,
} from './fileSearchFilters';

describe('fileSearchFilters', () => {
    it('parses an empty query string into the empty filters', () => {
        expect(parseFileSearchFilters(new URLSearchParams())).toEqual(emptyFileSearchFilters);
    });

    it('parses every filter from the url and drops unknown or malformed values', () => {
        const params = new URLSearchParams(
            'q=foto&kind=image&kind=folder&kind=bogus&modified_from=2026-01-01&modified_to=ontem&size=huge&tier=cold&starred=true&sort=size&order=asc'
        );

        expect(parseFileSearchFilters(params)).toEqual({
            kinds: ['folder', 'image'],
            modifiedFrom: '2026-01-01',
            modifiedTo: '',
            sizePreset: 'huge',
            tier: 'cold',
            onlyStarred: true,
            sort: 'size',
            order: 'asc',
        });
    });

    it('falls back to relevance for an unknown sort', () => {
        expect(parseFileSearchFilters(new URLSearchParams('sort=color')).sort).toBe('relevance');
    });

    it('writes only non-default filters and keeps unrelated params', () => {
        const written = writeFileSearchFilters(new URLSearchParams('q=foto&kind=audio'), {
            ...emptyFileSearchFilters,
            kinds: ['image', 'video'],
            onlyStarred: true,
            sort: 'name',
            order: 'desc',
        });

        expect(written.get('q')).toBe('foto');
        expect(written.getAll('kind')).toEqual(['image', 'video']);
        expect(written.get('starred')).toBe('true');
        expect(written.get('sort')).toBe('name');
        expect(written.get('order')).toBe('desc');
        expect(written.has('tier')).toBe(false);
    });

    it('removes every filter param when written with the empty filters', () => {
        const written = writeFileSearchFilters(
            new URLSearchParams('q=foto&kind=image&size=small&tier=hot&sort=name'),
            emptyFileSearchFilters
        );

        expect(written.toString()).toBe('q=foto');
    });

    it('writes period, size and tier', () => {
        const written = writeFileSearchFilters(new URLSearchParams(), {
            ...emptyFileSearchFilters,
            modifiedFrom: '2026-01-01',
            modifiedTo: '2026-01-31',
            sizePreset: 'medium',
            tier: 'hot',
        });

        expect(written.get('modified_from')).toBe('2026-01-01');
        expect(written.get('modified_to')).toBe('2026-01-31');
        expect(written.get('size')).toBe('medium');
        expect(written.get('tier')).toBe('hot');
    });

    it('counts each active filter group once', () => {
        expect(countActiveFileSearchFilters(emptyFileSearchFilters)).toBe(0);
        expect(
            countActiveFileSearchFilters({
                ...emptyFileSearchFilters,
                kinds: ['image', 'audio'],
                modifiedFrom: '2026-01-01',
                modifiedTo: '2026-02-01',
                onlyStarred: true,
                sort: 'size',
            })
        ).toBe(4);
    });

    it('maps the empty filters to refinements without any value', () => {
        expect(toFileSearchRefinements(emptyFileSearchFilters)).toEqual({
            kinds: undefined,
            modifiedFrom: undefined,
            modifiedTo: undefined,
            tier: undefined,
            starred: undefined,
            sort: undefined,
            order: undefined,
        });
    });

    it.each([
        ['small', { maxSize: 1048575 }],
        ['medium', { minSize: 1048576, maxSize: 104857599 }],
        ['large', { minSize: 104857600, maxSize: 1073741823 }],
        ['huge', { minSize: 1073741824 }],
    ] as const)('maps the %s size preset to byte bounds', (sizePreset, bounds) => {
        expect(toFileSearchRefinements({ ...emptyFileSearchFilters, sizePreset })).toMatchObject(
            bounds
        );
    });

    it('maps explicit sort, order, tier and starred to refinements', () => {
        expect(
            toFileSearchRefinements({
                ...emptyFileSearchFilters,
                kinds: ['image'],
                tier: 'cold',
                onlyStarred: true,
                sort: 'modified',
                order: 'asc',
            })
        ).toMatchObject({
            kinds: ['image'],
            tier: 'cold',
            starred: true,
            sort: 'modified',
            order: 'asc',
        });
    });

    it('never sends an order for relevance', () => {
        expect(
            toFileSearchRefinements({ ...emptyFileSearchFilters, sort: 'relevance', order: 'asc' })
                .order
        ).toBeUndefined();
    });
});
