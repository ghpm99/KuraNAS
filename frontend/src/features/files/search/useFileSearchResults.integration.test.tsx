import type { ReactNode } from 'react';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import useFileSearchResults from './useFileSearchResults';

const mockGet = jest.fn();

jest.mock('@/service', () => ({
    apiBase: { get: (...args: unknown[]) => mockGet(...args) },
}));

const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
        {children}
    </QueryClientProvider>
);

describe('file search seam', () => {
    beforeEach(() => jest.clearAllMocks());

    it('sends the exact query the backend decodes for a recursive folder search', async () => {
        mockGet.mockResolvedValue({ data: { items: [], pagination: { page: 1, hasNext: false } } });

        renderHook(
            () => useFileSearchResults({ query: 'relatorio', parentId: 7, isRecursive: true }),
            { wrapper }
        );

        await waitFor(() => expect(mockGet).toHaveBeenCalled());
        expect(mockGet).toHaveBeenCalledWith('/files/search', {
            params: {
                q: 'relatorio',
                parent_id: 7,
                recursive: true,
                page: 1,
                page_size: 100,
            },
            paramsSerializer: { indexes: null },
        });
    });

    it('sends a direct-children search with recursive=false', async () => {
        mockGet.mockResolvedValue({ data: { items: [], pagination: { page: 1, hasNext: false } } });

        renderHook(
            () => useFileSearchResults({ query: 'relatorio', parentId: 7, isRecursive: false }),
            { wrapper }
        );

        await waitFor(() => expect(mockGet).toHaveBeenCalled());
        expect(mockGet.mock.calls[0]?.[1]).toEqual({
            params: expect.objectContaining({ parent_id: 7, recursive: false }),
            paramsSerializer: { indexes: null },
        });
    });

    it('sends no parent_id for the global search', async () => {
        mockGet.mockResolvedValue({ data: { items: [], pagination: { page: 1, hasNext: false } } });

        renderHook(() => useFileSearchResults({ query: 'relatorio', isRecursive: true }), {
            wrapper,
        });

        await waitFor(() => expect(mockGet).toHaveBeenCalled());
        expect(mockGet.mock.calls[0]?.[1].params.parent_id).toBeUndefined();
    });

    it('sends every refinement with the exact names the backend decodes', async () => {
        mockGet.mockResolvedValue({ data: { items: [], pagination: { page: 1, hasNext: false } } });

        renderHook(
            () =>
                useFileSearchResults({
                    query: 'relatorio',
                    isRecursive: true,
                    refinements: {
                        kinds: ['image', 'folder'],
                        modifiedFrom: '2026-01-01',
                        modifiedTo: '2026-01-31',
                        minSize: 1048576,
                        maxSize: 104857599,
                        tier: 'cold',
                        starred: true,
                        sort: 'size',
                        order: 'asc',
                    },
                }),
            { wrapper }
        );

        await waitFor(() => expect(mockGet).toHaveBeenCalled());
        expect(mockGet).toHaveBeenCalledWith('/files/search', {
            params: {
                q: 'relatorio',
                parent_id: undefined,
                recursive: true,
                page: 1,
                page_size: 100,
                kind: ['image', 'folder'],
                modified_from: '2026-01-01',
                modified_to: '2026-01-31',
                min_size: 1048576,
                max_size: 104857599,
                tier: 'cold',
                starred: true,
                sort: 'size',
                order: 'asc',
            },
            paramsSerializer: { indexes: null },
        });
    });
});
