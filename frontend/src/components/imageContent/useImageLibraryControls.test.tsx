import { act, renderHook } from '@testing-library/react';
import type { ReactNode } from 'react';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { parseImageLibraryView } from './imageLibraryView';
import { useImageLibraryControls } from './useImageLibraryControls';

const setup = (initialRoute: string) => {
    const wrapper = ({ children }: { children: ReactNode }) => (
        <MemoryRouter initialEntries={[initialRoute]}>{children}</MemoryRouter>
    );
    return renderHook(
        () => {
            const location = useLocation();
            const searchParams = new URLSearchParams(location.search);
            const view = parseImageLibraryView('library', searchParams);
            return { controls: useImageLibraryControls(view), searchParams, location };
        },
        { wrapper }
    );
};

describe('useImageLibraryControls', () => {
    it('writes and clears the search term in the URL', () => {
        const { result } = setup('/images?before=2026-04-01&image=3');

        act(() => result.current.controls.setNameQuery('  beach '));
        expect(result.current.searchParams.get('q')).toBe('beach');
        expect(result.current.searchParams.get('before')).toBeNull();
        expect(result.current.searchParams.get('image')).toBeNull();

        act(() => result.current.controls.setNameQuery(''));
        expect(result.current.searchParams.has('q')).toBe(false);
    });

    it('writes the period bounds', () => {
        const { result } = setup('/images');

        act(() => result.current.controls.setTakenFrom('2026-01-01'));
        act(() => result.current.controls.setTakenTo('2026-02-01'));
        expect(result.current.searchParams.get('from')).toBe('2026-01-01');
        expect(result.current.searchParams.get('to')).toBe('2026-02-01');

        act(() => result.current.controls.setTakenFrom(''));
        expect(result.current.searchParams.has('from')).toBe(false);
    });

    it('toggles formats as repeated parameters', () => {
        const { result } = setup('/images');

        act(() => result.current.controls.toggleFormat('jpg'));
        act(() => result.current.controls.toggleFormat('png'));
        expect(result.current.searchParams.getAll('format')).toEqual(['jpg', 'png']);

        act(() => result.current.controls.toggleFormat('jpg'));
        expect(result.current.searchParams.getAll('format')).toEqual(['png']);
    });

    it('sets and clears the camera filter', () => {
        const { result } = setup('/images');

        act(() => result.current.controls.setCamera(' Canon EOS R5 '));
        expect(result.current.searchParams.get('camera')).toBe('Canon EOS R5');

        act(() => result.current.controls.setCamera(''));
        expect(result.current.searchParams.has('camera')).toBe(false);
    });

    it('changes sort and flips the order', () => {
        const { result } = setup('/images?order=asc');

        act(() => result.current.controls.setSort('size'));
        expect(result.current.searchParams.get('sort')).toBe('size');
        expect(result.current.searchParams.has('order')).toBe(false);

        act(() => result.current.controls.toggleSortOrder());
        expect(result.current.searchParams.get('order')).toBe('asc');

        act(() => result.current.controls.setSort('taken_at'));
        expect(result.current.searchParams.has('sort')).toBe(false);
    });

    it('clears only the user filters', () => {
        const { result } = setup(
            '/images?q=a&from=2026-01-01&to=2026-02-01&format=jpg&camera=Sony&sort=name'
        );

        act(() => result.current.controls.clearUserFilters());

        expect(result.current.location.search).toBe('?sort=name');
    });

    it('jumps to the first instant after a month and returns to the latest', () => {
        const { result } = setup('/images?image=3&imagePath=%2Fa');

        act(() => result.current.controls.jumpToMonth(2026, 3));
        expect(result.current.searchParams.get('before')).toBe('2026-04-01');
        expect(result.current.searchParams.has('image')).toBe(false);
        expect(result.current.searchParams.has('imagePath')).toBe(false);

        act(() => result.current.controls.clearJump());
        expect(result.current.searchParams.has('before')).toBe(false);
    });

    it('selects folders and albums and opens or closes the viewer without losing the jump', () => {
        const { result } = setup('/images?before=2026-04-01');

        act(() => result.current.controls.openImageParam(5));
        expect(result.current.searchParams.get('image')).toBe('5');
        expect(result.current.searchParams.get('before')).toBe('2026-04-01');

        act(() => result.current.controls.closeImageParam());
        expect(result.current.searchParams.has('image')).toBe(false);

        act(() => result.current.controls.selectFolder('/photos'));
        expect(result.current.searchParams.get('folder')).toBe('/photos');
        act(() => result.current.controls.selectFolder(null));
        expect(result.current.searchParams.has('folder')).toBe(false);

        act(() => result.current.controls.selectAlbum('memes'));
        expect(result.current.searchParams.get('album')).toBe('memes');
        act(() => result.current.controls.selectAlbum(null));
        expect(result.current.searchParams.has('album')).toBe(false);
    });

    it('opens a user album in the URL and leaves it when going back to the albums list', () => {
        const { result } = setup('/images?album=memes&image=3');

        act(() => result.current.controls.selectUserAlbum(12));
        expect(result.current.searchParams.get('userAlbum')).toBe('12');
        expect(result.current.searchParams.has('album')).toBe(false);
        expect(result.current.searchParams.has('image')).toBe(false);

        act(() => result.current.controls.selectAlbum(null));
        expect(result.current.searchParams.has('userAlbum')).toBe(false);

        act(() => result.current.controls.selectUserAlbum(12));
        act(() => result.current.controls.selectUserAlbum(null));
        expect(result.current.searchParams.has('userAlbum')).toBe(false);
    });
});
