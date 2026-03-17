import { renderHook } from '@testing-library/react';
import { useGlobalSearch, GlobalSearchContext } from './useGlobalSearch';
import React from 'react';

describe('useGlobalSearch', () => {
    it('throws when used outside a provider', () => {
        jest.spyOn(console, 'error').mockImplementation(() => {});

        expect(() => {
            renderHook(() => useGlobalSearch());
        }).toThrow('useGlobalSearch must be used within a GlobalSearchProvider');

        (console.error as jest.Mock).mockRestore();
    });

    it('returns the context value when inside a provider', () => {
        const contextValue = {
            openSearch: jest.fn(),
            shortcut: 'Ctrl+K',
        };

        const wrapper = ({ children }: { children: React.ReactNode }) =>
            React.createElement(GlobalSearchContext.Provider, { value: contextValue }, children);

        const { result } = renderHook(() => useGlobalSearch(), { wrapper });

        expect(result.current.openSearch).toBe(contextValue.openSearch);
        expect(result.current.shortcut).toBe('Ctrl+K');
    });
});
