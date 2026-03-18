import { renderHook } from '@testing-library/react';
import { useGlobalSearch } from './useGlobalSearch';

describe('useGlobalSearch', () => {
    it('throws when used outside a provider', () => {
        jest.spyOn(console, 'error').mockImplementation(() => {});

        expect(() => {
            renderHook(() => useGlobalSearch());
        }).toThrow('useGlobalSearch must be used within a GlobalSearchProvider');

        (console.error as jest.Mock).mockRestore();
    });
});
