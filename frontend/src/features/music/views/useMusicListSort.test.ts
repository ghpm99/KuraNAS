import { act, renderHook } from '@testing-library/react';
import { getMusicListSortStorageKey } from './musicListSortPreference';
import { useMusicListSort } from './useMusicListSort';

describe('useMusicListSort', () => {
    beforeEach(() => {
        window.localStorage.clear();
    });

    it('starts from the default sort when nothing is stored', () => {
        const { result } = renderHook(() => useMusicListSort('artists'));

        expect(result.current.listSort).toEqual({ sort: 'tracks', order: 'desc' });
    });

    it('applies the natural order of the chosen field and persists it', () => {
        const { result } = renderHook(() => useMusicListSort('albums'));

        act(() => result.current.changeField('name'));
        expect(result.current.listSort).toEqual({ sort: 'name', order: 'asc' });

        act(() => result.current.toggleOrder());
        expect(result.current.listSort).toEqual({ sort: 'name', order: 'desc' });
        expect(window.localStorage.getItem(getMusicListSortStorageKey('albums'))).toBe(
            JSON.stringify({ sort: 'name', order: 'desc' })
        );

        act(() => result.current.changeField('recent'));
        expect(result.current.listSort).toEqual({ sort: 'recent', order: 'desc' });

        act(() => result.current.toggleOrder());
        expect(result.current.listSort).toEqual({ sort: 'recent', order: 'asc' });
    });

    it('restores the persisted sort on mount', () => {
        window.localStorage.setItem(
            getMusicListSortStorageKey('folders'),
            JSON.stringify({ sort: 'recent', order: 'asc' })
        );

        const { result } = renderHook(() => useMusicListSort('folders'));

        expect(result.current.listSort).toEqual({ sort: 'recent', order: 'asc' });
    });
});
