import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import { apiBase } from '@/service';
import { useAnalyticsBlockQuery } from './useAnalyticsBlockQuery';

jest.mock('@/service', () => ({ apiBase: { get: jest.fn() } }));

const mockedGet = apiBase.get as jest.Mock;

const fetchStorageBlock = async () => (await apiBase.get('/analytics/storage')).data;

const renderBlockHook = () => {
    const queryClient = new QueryClient();
    const wrapper = ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
    return renderHook(
        () =>
            useAnalyticsBlockQuery({
                queryKey: ['analytics', 'storage-test'],
                fetchBlock: fetchStorageBlock,
            }),
        { wrapper }
    );
};

describe('useAnalyticsBlockQuery', () => {
    beforeEach(() => mockedGet.mockReset());

    it('starts pending and then exposes the fetched data', async () => {
        mockedGet.mockResolvedValue({ data: { used: 10 } });
        const { result } = renderBlockHook();

        expect(result.current.status).toBe('pending');
        await waitFor(() => expect(result.current.status).toBe('success'));
        expect(result.current.data).toEqual({ used: 10 });
    });

    it('exposes the server message on error and never a zero fallback', async () => {
        mockedGet.mockRejectedValue({ response: { data: { error: 'Banco indisponível' } } });
        const { result } = renderBlockHook();

        await waitFor(() => expect(result.current.status).toBe('error'));
        expect(result.current.errorMessage).toBe('Banco indisponível');
        expect(result.current.data).toBeUndefined();
    });

    it('refetches on retry and recovers', async () => {
        mockedGet.mockRejectedValueOnce(new Error('boom'));
        mockedGet.mockResolvedValueOnce({ data: { used: 3 } });
        const { result } = renderBlockHook();

        await waitFor(() => expect(result.current.status).toBe('error'));
        act(() => result.current.retry());

        await waitFor(() => expect(result.current.status).toBe('success'));
        expect(result.current.data).toEqual({ used: 3 });
        expect(mockedGet).toHaveBeenCalledTimes(2);
    });
});
