import type { ReactNode } from 'react';
import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import useDocumentSearchResults from './useDocumentSearchResults';

const mockGet = jest.fn();

jest.mock('@/service', () => ({
    apiBase: { get: (...args: unknown[]) => mockGet(...args) },
}));

const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
        {children}
    </QueryClientProvider>
);

const pageOf = (fileId: number, page: number, hasNext: boolean) => ({
    data: {
        items: [{ file_id: fileId, name: `doc-${fileId}.txt`, path: `/doc-${fileId}.txt` }],
        pagination: { page, page_size: 20, has_next: hasNext, has_prev: page > 1 },
    },
});

describe('useDocumentSearchResults seam', () => {
    beforeEach(() => jest.clearAllMocks());

    it('sends the exact query the backend decodes', async () => {
        mockGet.mockResolvedValue(pageOf(1, 1, false));

        renderHook(() => useDocumentSearchResults({ query: 'contrato', isEnabled: true }), {
            wrapper,
        });

        await waitFor(() => expect(mockGet).toHaveBeenCalled());
        expect(mockGet).toHaveBeenCalledWith('/documents/search', {
            params: { q: 'contrato', page: 1, page_size: 20 },
            signal: expect.anything(),
        });
    });

    it('does not request anything while disabled or without a query', () => {
        renderHook(() => useDocumentSearchResults({ query: 'contrato', isEnabled: false }), {
            wrapper,
        });
        renderHook(() => useDocumentSearchResults({ query: '', isEnabled: true }), { wrapper });

        expect(mockGet).not.toHaveBeenCalled();
    });

    it('accumulates pages through has_next', async () => {
        mockGet
            .mockResolvedValueOnce(pageOf(1, 1, true))
            .mockResolvedValueOnce(pageOf(2, 2, false));

        const { result } = renderHook(
            () => useDocumentSearchResults({ query: 'contrato', isEnabled: true }),
            { wrapper }
        );

        await waitFor(() => expect(result.current.items).toHaveLength(1));
        expect(result.current.hasNextPage).toBe(true);

        act(() => result.current.fetchNextPage());

        await waitFor(() => expect(result.current.items).toHaveLength(2));
        expect(mockGet).toHaveBeenLastCalledWith(
            '/documents/search',
            expect.objectContaining({ params: { q: 'contrato', page: 2, page_size: 20 } })
        );
        expect(result.current.hasNextPage).toBe(false);
    });

    it('exposes the backend error message and retries', async () => {
        mockGet.mockRejectedValueOnce({ response: { data: { error: 'indisponivel' } } });
        mockGet.mockResolvedValueOnce(pageOf(1, 1, false));

        const { result } = renderHook(
            () => useDocumentSearchResults({ query: 'contrato', isEnabled: true }),
            { wrapper }
        );

        await waitFor(() => expect(result.current.status).toBe('error'));
        expect(result.current.errorMessage).toBe('indisponivel');

        act(() => result.current.retry());

        await waitFor(() => expect(result.current.status).toBe('success'));
    });
});
