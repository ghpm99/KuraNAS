import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider, onlineManager } from '@tanstack/react-query';
import { AxiosError } from 'axios';
import { getServerHealth } from '@/service/health';
import ConnectionBanner from './ConnectionBanner';

jest.mock('@/service/health', () => ({ getServerHealth: jest.fn() }));

const mockedGetServerHealth = getServerHealth as jest.Mock;

const renderBanner = () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const invalidateSpy = jest.spyOn(queryClient, 'invalidateQueries');
    render(
        <QueryClientProvider client={queryClient}>
            <ConnectionBanner />
        </QueryClientProvider>
    );
    return { queryClient, invalidateSpy };
};

const failQueryWithNetworkError = async (queryClient: QueryClient) => {
    await act(async () => {
        await queryClient
            .fetchQuery({
                queryKey: ['network-failure'],
                queryFn: () => Promise.reject(new AxiosError('Network Error')),
            })
            .catch(() => undefined);
    });
};

describe('layout/ConnectionBanner', () => {
    beforeEach(() => {
        mockedGetServerHealth.mockReset();
        onlineManager.setOnline(true);
    });

    afterEach(() => {
        onlineManager.setOnline(true);
        jest.useRealTimers();
    });

    it('renders nothing while the connection is healthy', () => {
        renderBanner();

        expect(screen.queryByRole('status')).not.toBeInTheDocument();
    });

    it('shows the banner when the browser goes offline and hides it when back online', () => {
        renderBanner();

        act(() => onlineManager.setOnline(false));
        expect(screen.getByRole('status')).toHaveTextContent('CONNECTION_BANNER_UNAVAILABLE');

        act(() => onlineManager.setOnline(true));
        expect(screen.queryByRole('status')).not.toBeInTheDocument();
    });

    it('shows the banner on a network failure and recovers through retry now', async () => {
        const { queryClient, invalidateSpy } = renderBanner();
        mockedGetServerHealth.mockRejectedValueOnce(new Error('down'));

        await failQueryWithNetworkError(queryClient);
        expect(screen.getByRole('status')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'CONNECTION_RETRY_NOW' }));
        await waitFor(() => expect(mockedGetServerHealth).toHaveBeenCalledTimes(1));
        expect(screen.getByRole('status')).toBeInTheDocument();

        mockedGetServerHealth.mockResolvedValueOnce({ status: 'ok', service: 'kuranas' });
        fireEvent.click(screen.getByRole('button', { name: 'CONNECTION_RETRY_NOW' }));

        await waitFor(() => expect(screen.queryByRole('status')).not.toBeInTheDocument());
        expect(invalidateSpy).toHaveBeenCalled();
    });

    it('ignores failures that carry an http response', async () => {
        const { queryClient } = renderBanner();
        const serverError = new AxiosError('boom');
        serverError.response = { status: 500 } as AxiosError['response'];

        await act(async () => {
            await queryClient
                .fetchQuery({ queryKey: ['server-failure'], queryFn: () => Promise.reject(serverError) })
                .catch(() => undefined);
        });

        expect(screen.queryByRole('status')).not.toBeInTheDocument();
    });

    it('polls the health endpoint with backoff and hides when the server returns', async () => {
        jest.useFakeTimers();
        const { queryClient } = renderBanner();
        mockedGetServerHealth
            .mockRejectedValueOnce(new Error('down'))
            .mockResolvedValueOnce({ status: 'ok', service: 'kuranas' });

        await failQueryWithNetworkError(queryClient);
        expect(screen.getByRole('status')).toBeInTheDocument();

        await act(async () => {
            await jest.advanceTimersByTimeAsync(2000);
        });
        expect(mockedGetServerHealth).toHaveBeenCalledTimes(1);
        expect(screen.getByRole('status')).toBeInTheDocument();

        await act(async () => {
            await jest.advanceTimersByTimeAsync(4000);
        });
        expect(mockedGetServerHealth).toHaveBeenCalledTimes(2);
        expect(screen.queryByRole('status')).not.toBeInTheDocument();
    });
});
