import type { ReactNode } from 'react';
import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import useFileSearchResults from './useFileSearchResults';

const mockSearchFiles = jest.fn();

jest.mock('@/service/files', () => ({
    searchFiles: (...args: unknown[]) => mockSearchFiles(...args),
}));

const createWrapper = () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
};

describe('useFileSearchResults', () => {
    beforeEach(() => jest.clearAllMocks());

    it('stays idle and never calls the backend without a query', () => {
        const { result } = renderHook(
            () => useFileSearchResults({ query: '', isRecursive: true }),
            { wrapper: createWrapper() }
        );

        expect(result.current.items).toEqual([]);
        expect(result.current.hasNextPage).toBe(false);
        expect(mockSearchFiles).not.toHaveBeenCalled();
    });

    it('survives a failing backend', async () => {
        mockSearchFiles.mockRejectedValue(new Error('offline'));
        const { result } = renderHook(
            () => useFileSearchResults({ query: 'foto', isRecursive: true }),
            { wrapper: createWrapper() }
        );

        await waitFor(() => expect(result.current.status).toBe('error'));
        expect(result.current.items).toEqual([]);
    });

    it('exposes the backend error message and retries the search', async () => {
        mockSearchFiles.mockRejectedValueOnce({ response: { data: { error: 'Busca indisponível' } } });
        mockSearchFiles.mockResolvedValue({
            items: [{ id: 1, name: 'foto.png' }],
            pagination: { page: 1, pageSize: 100, hasNext: false },
        });
        const { result } = renderHook(
            () => useFileSearchResults({ query: 'foto', isRecursive: true }),
            { wrapper: createWrapper() }
        );

        await waitFor(() => expect(result.current.errorMessage).toBe('Busca indisponível'));
        act(() => result.current.retry());

        await waitFor(() => expect(result.current.items).toHaveLength(1));
    });

    it('survives a payload without items or pagination', async () => {
        mockSearchFiles.mockResolvedValue({});
        const { result } = renderHook(
            () => useFileSearchResults({ query: 'foto', isRecursive: true }),
            { wrapper: createWrapper() }
        );

        await waitFor(() => expect(result.current.status).toBe('success'));
        expect(result.current.items).toEqual([]);
        expect(result.current.hasNextPage).toBe(false);
    });

    it('searches the folder scope and walks the next pages', async () => {
        mockSearchFiles
            .mockResolvedValueOnce({
                items: [{ id: 1 }],
                pagination: { page: 1, hasNext: true },
            })
            .mockResolvedValueOnce({
                items: [{ id: 2 }],
                pagination: { page: 2, hasNext: false },
            });
        const { result } = renderHook(
            () => useFileSearchResults({ query: 'foto', parentId: 7, isRecursive: false }),
            { wrapper: createWrapper() }
        );

        await waitFor(() => expect(result.current.items).toHaveLength(1));
        expect(result.current.hasNextPage).toBe(true);
        expect(mockSearchFiles).toHaveBeenCalledWith({
            q: 'foto',
            parentId: 7,
            recursive: false,
            page: 1,
            pageSize: 100,
        });

        result.current.fetchNextPage();

        await waitFor(() => expect(result.current.items).toHaveLength(2));
        expect(result.current.hasNextPage).toBe(false);
        expect(mockSearchFiles).toHaveBeenLastCalledWith(expect.objectContaining({ page: 2 }));
    });
});
