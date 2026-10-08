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
});
