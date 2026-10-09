import {
    buildJumpBeforeDate,
    getDefaultSortOrder,
    imageAlbumPresets,
    isKeysetOrdering,
    parseImageLibraryView,
} from './imageLibraryView';

const now = new Date('2026-06-15T12:00:00Z');
const parse = (section: Parameters<typeof parseImageLibraryView>[0], query = '') =>
    parseImageLibraryView(section, new URLSearchParams(query), now);

describe('imageLibraryView', () => {
    it('applies no filter on the library section', () => {
        const view = parse('library');

        expect(view.filters).toEqual({
            nameQuery: '',
            categories: [],
            isStarredOnly: false,
            formats: [],
            camera: '',
            takenFrom: '',
            takenTo: '',
            folder: '',
        });
        expect(view.ordering).toEqual({ sort: 'taken_at', order: 'desc' });
        expect(view.isKeyset).toBe(true);
        expect(view.hasUserFilters).toBe(false);
    });

    it.each([
        ['captures', ['capture', 'screenshot_app']],
        ['photos', ['photo', 'landscape', 'portrait']],
    ] as const)('maps the %s section to server categories', (section, categories) => {
        expect(parse(section).filters.categories).toEqual(categories);
    });

    it('maps recent to the last 30 days and lets the user override the start', () => {
        expect(parse('recent').filters.takenFrom).toBe('2026-05-16');
        const overridden = parse('recent', 'from=2026-06-01');
        expect(overridden.filters.takenFrom).toBe('2026-06-01');
        expect(overridden.hasUserFilters).toBe(true);
    });

    it('maps favorites to the starred filter', () => {
        expect(parse('favorites').filters.isStarredOnly).toBe(true);
    });

    it('reads the camera filter and counts it as a user filter', () => {
        const view = parseImageLibraryView(
            'library',
            new URLSearchParams('camera=%20Canon%20EOS%20')
        );

        expect(view.filters.camera).toBe('Canon EOS');
        expect(view.hasUserFilters).toBe(true);
    });

    it('reads search, period, formats and ignores invalid values', () => {
        const view = parse(
            'library',
            'q=%20beach%20&from=2026-01-01&to=nope&format=jpg&format=exe&sort=bogus&order=sideways'
        );

        expect(view.filters.nameQuery).toBe('beach');
        expect(view.filters.takenFrom).toBe('2026-01-01');
        expect(view.filters.takenTo).toBe('');
        expect(view.filters.formats).toEqual(['jpg']);
        expect(view.ordering).toEqual({ sort: 'taken_at', order: 'desc' });
        expect(view.hasUserFilters).toBe(true);
    });

    it('uses ascending order by default for name and disables the keyset', () => {
        const view = parse('library', 'sort=name&before=2026-04-01');

        expect(view.ordering).toEqual({ sort: 'name', order: 'asc' });
        expect(view.isKeyset).toBe(false);
        expect(view.takenBefore).toBe('');
        expect(getDefaultSortOrder('size')).toBe('desc');
        expect(isKeysetOrdering({ sort: 'taken_at', order: 'asc' })).toBe(false);
    });

    it('keeps the jump date only for keyset ordering', () => {
        expect(parse('library', 'before=2026-04-01').takenBefore).toBe('2026-04-01');
        expect(parse('library', 'before=garbage').takenBefore).toBe('');
    });

    it('selects folders and albums only inside their own section', () => {
        expect(parse('folders', 'folder=%2Fa%2Fb').filters.folder).toBe('/a/b');
        expect(parse('library', 'folder=%2Fa%2Fb').filters.folder).toBe('');

        const albumView = parse('albums', 'album=memes');
        expect(albumView.selectedAlbum?.id).toBe('memes');
        expect(albumView.filters.categories).toEqual(['meme']);
        expect(parse('albums', 'album=unknown').selectedAlbum).toBeNull();
        expect(parse('photos', 'album=memes').selectedAlbum).toBeNull();
    });

    it('reads the user album id only inside the albums section and without a preset', () => {
        expect(parse('albums', 'userAlbum=7').userAlbumId).toBe(7);
        expect(parse('albums', 'userAlbum=abc').userAlbumId).toBeNull();
        expect(parse('albums', 'userAlbum=-3').userAlbumId).toBeNull();
        expect(parse('albums', 'userAlbum=1.5').userAlbumId).toBeNull();
        expect(parse('albums', '').userAlbumId).toBeNull();
        expect(parse('photos', 'userAlbum=7').userAlbumId).toBeNull();
        expect(parse('albums', 'album=memes&userAlbum=7').userAlbumId).toBeNull();
    });

    it('exposes the five album presets backed by server categories', () => {
        expect(imageAlbumPresets.map((preset) => preset.id)).toEqual([
            'documents',
            'memes',
            'art',
            'landscapes',
            'portraits',
        ]);
        expect(imageAlbumPresets[0]!.categories).toEqual(['document', 'receipt']);
    });

    it('builds the first instant after a month as a date-only value', () => {
        expect(buildJumpBeforeDate(2026, 3)).toBe('2026-04-01');
        expect(buildJumpBeforeDate(2025, 12)).toBe('2026-01-01');
    });
});
