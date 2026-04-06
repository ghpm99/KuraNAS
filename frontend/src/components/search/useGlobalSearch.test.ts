import { renderHook } from '@testing-library/react';
import { createElement } from 'react';
import { useGlobalSearch, GlobalSearchContext, type GlobalSearchContextValue } from './useGlobalSearch';

describe('useGlobalSearch', () => {
    it('throws when used outside a provider', () => {
        jest.spyOn(console, 'error').mockImplementation(() => {});

        expect(() => {
            renderHook(() => useGlobalSearch());
        }).toThrow('useGlobalSearch must be used within a GlobalSearchProvider');

        (console.error as jest.Mock).mockRestore();
    });

    it('returns context value when inside provider', () => {
        const value: GlobalSearchContextValue = { openSearch: jest.fn(), shortcut: 'Ctrl+K' };
        const wrapper = ({ children }: { children: React.ReactNode }) =>
            createElement(GlobalSearchContext.Provider, { value }, children);

        const { result } = renderHook(() => useGlobalSearch(), { wrapper });
        expect(result.current.shortcut).toBe('Ctrl+K');
    });
});
