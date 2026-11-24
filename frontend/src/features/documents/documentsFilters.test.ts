import {
    countActiveDocumentsFilters,
    defaultDocumentsSort,
    emptyDocumentsFilters,
    parseDocumentsFilters,
    parseDocumentsSort,
    toDocumentFilters,
    toDocumentSort,
    writeDocumentsFilters,
    writeDocumentsSort,
    type DocumentsFilterState,
} from './documentsFilters';

const combinedFilters: DocumentsFilterState = {
    format: 'pdf',
    modifiedFrom: '2026-01-01',
    modifiedTo: '2026-02-01',
    favoritesOnly: true,
    indexState: 'failed',
};

describe('documentsFilters', () => {
    it('parses an empty query string into the empty filters', () => {
        expect(parseDocumentsFilters(new URLSearchParams())).toEqual(emptyDocumentsFilters);
        expect(parseDocumentsSort(new URLSearchParams())).toBeNull();
    });

    it('round-trips every filter through the URL', () => {
        const params = writeDocumentsFilters(new URLSearchParams(), combinedFilters);

        expect(params.toString()).toBe(
            'format=pdf&modified_from=2026-01-01&modified_to=2026-02-01&favorites_only=true&index_state=failed'
        );
        expect(parseDocumentsFilters(params)).toEqual(combinedFilters);
    });

    it('drops filters that were cleared and keeps unrelated params', () => {
        const params = writeDocumentsFilters(
            new URLSearchParams('format=pdf&sort=name&order=asc&folder=3'),
            emptyDocumentsFilters
        );

        expect(params.toString()).toBe('sort=name&order=asc&folder=3');
    });

    it('falls back to defaults for invalid values', () => {
        const params = new URLSearchParams(
            'format=exe&modified_from=ontem&modified_to=2026-13&favorites_only=yes&index_state=lost&sort=color&order=up'
        );

        expect(parseDocumentsFilters(params)).toEqual(emptyDocumentsFilters);
        expect(parseDocumentsSort(params)).toBeNull();
    });

    it('round-trips the sort and writes both sort and order', () => {
        const params = writeDocumentsSort(new URLSearchParams('format=pdf'), {
            sort: 'size',
            order: 'asc',
        });

        expect(params.toString()).toBe('format=pdf&sort=size&order=asc');
        expect(parseDocumentsSort(params)).toEqual({ sort: 'size', order: 'asc' });
    });

    it('uses the natural order of each sort field when the URL has no order', () => {
        expect(parseDocumentsSort(new URLSearchParams('sort=name'))).toEqual({
            sort: 'name',
            order: 'asc',
        });
        expect(parseDocumentsSort(new URLSearchParams('sort=modified_at&order=bad'))).toEqual(
            defaultDocumentsSort
        );
        expect(parseDocumentsSort(new URLSearchParams('sort=created_at'))).toEqual({
            sort: 'created_at',
            order: 'desc',
        });
    });

    it('counts a period as a single active filter', () => {
        expect(countActiveDocumentsFilters(emptyDocumentsFilters)).toBe(0);
        expect(countActiveDocumentsFilters(combinedFilters)).toBe(4);
        expect(
            countActiveDocumentsFilters({ ...emptyDocumentsFilters, modifiedTo: '2026-02-01' })
        ).toBe(1);
    });

    it('maps to service filters leaving cleared ones undefined', () => {
        expect(toDocumentFilters(emptyDocumentsFilters)).toEqual({
            format: undefined,
            modifiedFrom: undefined,
            modifiedTo: undefined,
            favoritesOnly: undefined,
            indexState: undefined,
        });
        expect(toDocumentFilters(combinedFilters)).toEqual({
            format: 'pdf',
            modifiedFrom: '2026-01-01',
            modifiedTo: '2026-02-01',
            favoritesOnly: true,
            indexState: 'failed',
        });
        expect(toDocumentSort(defaultDocumentsSort)).toEqual({
            sort: 'modified_at',
            order: 'desc',
        });
    });
});
