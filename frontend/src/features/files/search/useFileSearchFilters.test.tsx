import type { ReactNode } from 'react';
import { act, renderHook } from '@testing-library/react';
import { MemoryRouter, useLocation } from 'react-router-dom';
import useFileSearchFilters from './useFileSearchFilters';

const renderFiltersHook = (initialUrl: string) =>
    renderHook(() => ({ filters: useFileSearchFilters(), location: useLocation() }), {
        wrapper: ({ children }: { children: ReactNode }) => (
            <MemoryRouter initialEntries={[initialUrl]}>{children}</MemoryRouter>
        ),
    });

describe('useFileSearchFilters', () => {
    it('starts with the empty filters without any url params', () => {
        const { result } = renderFiltersHook('/files');

        expect(result.current.filters.filters.kinds).toEqual([]);
        expect(result.current.filters.filters.sort).toBe('relevance');
    });

    it('reads the filters from a shared url', () => {
        const { result } = renderFiltersHook('/files?q=foto&tier=cold&kind=image');

        expect(result.current.filters.filters.tier).toBe('cold');
        expect(result.current.filters.filters.kinds).toEqual(['image']);
    });

    it('writes filters to the url while keeping q', () => {
        const { result } = renderFiltersHook('/files?q=foto');

        act(() =>
            result.current.filters.setFilters({
                ...result.current.filters.filters,
                onlyStarred: true,
            })
        );

        expect(result.current.location.search).toBe('?q=foto&starred=true');
        expect(result.current.filters.filters.onlyStarred).toBe(true);
    });

    it('resets every filter and keeps q', () => {
        const { result } = renderFiltersHook('/files?q=foto&starred=true&kind=audio&sort=size');

        act(() => result.current.filters.resetFilters());

        expect(result.current.location.search).toBe('?q=foto');
    });
});
