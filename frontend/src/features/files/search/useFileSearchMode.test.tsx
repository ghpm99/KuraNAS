import type { ReactNode } from 'react';
import { act, renderHook } from '@testing-library/react';
import { MemoryRouter, useLocation } from 'react-router-dom';
import useFileSearchMode from './useFileSearchMode';

const renderModeHook = (initialUrl: string) => {
    const wrapper = ({ children }: { children: ReactNode }) => (
        <MemoryRouter initialEntries={[initialUrl]}>{children}</MemoryRouter>
    );
    return renderHook(() => ({ mode: useFileSearchMode(), location: useLocation() }), { wrapper });
};

describe('useFileSearchMode', () => {
    it('defaults to name mode without the in param', () => {
        const { result } = renderModeHook('/files?q=abc');

        expect(result.current.mode.searchMode).toBe('name');
    });

    it('reads content mode from the url', () => {
        const { result } = renderModeHook('/files?q=abc&in=content');

        expect(result.current.mode.searchMode).toBe('content');
    });

    it('writes and removes the in param while keeping the query', () => {
        const { result } = renderModeHook('/files?q=abc');

        act(() => result.current.mode.setSearchMode('content'));
        expect(result.current.location.search).toBe('?q=abc&in=content');

        act(() => result.current.mode.setSearchMode('name'));
        expect(result.current.location.search).toBe('?q=abc');
    });
});
