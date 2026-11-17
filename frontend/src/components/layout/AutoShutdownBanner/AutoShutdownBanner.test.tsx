import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import AutoShutdownBanner from './AutoShutdownBanner';
import { cancelAutoShutdownToday, getAutoShutdownStatus } from '@/service/autoShutdown';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, vars?: Record<string, string>) => (vars ? `${key}:${vars.seconds}` : key),
    }),
}));

jest.mock('@/service/autoShutdown', () => ({
    getAutoShutdownStatus: jest.fn(),
    cancelAutoShutdownToday: jest.fn(),
}));

const mockedGetStatus = getAutoShutdownStatus as jest.Mock;
const mockedCancel = cancelAutoShutdownToday as jest.Mock;

const idleStatus = {
    state: 'idle',
    shutdown_at: null,
    seconds_remaining: null,
    postponed_reason: '',
};
const warningStatus = {
    state: 'warning',
    shutdown_at: '2026-11-17T03:00:00Z',
    seconds_remaining: 42,
    postponed_reason: '',
};

const renderBanner = () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <AutoShutdownBanner />
        </QueryClientProvider>
    );
};

describe('layout/AutoShutdownBanner', () => {
    describe('with a mocked service', () => {
        beforeEach(() => {
            jest.resetAllMocks();
        });

        afterEach(() => {
            jest.useRealTimers();
        });

        it('renders nothing while idle', async () => {
            mockedGetStatus.mockResolvedValue(idleStatus);
            renderBanner();

            await waitFor(() => expect(mockedGetStatus).toHaveBeenCalled());
            expect(screen.queryByRole('alert')).not.toBeInTheDocument();
        });

        it('renders nothing for a partial warning payload', async () => {
            mockedGetStatus.mockResolvedValue({ state: 'warning' });
            renderBanner();

            await waitFor(() => expect(mockedGetStatus).toHaveBeenCalled());
            expect(screen.queryByRole('alert')).not.toBeInTheDocument();
        });

        it('shows the countdown and decrements it every second', async () => {
            jest.useFakeTimers();
            mockedGetStatus.mockResolvedValue(warningStatus);
            renderBanner();

            expect(await screen.findByRole('alert')).toHaveTextContent(
                'AUTO_SHUTDOWN_BANNER_MESSAGE:42'
            );
            expect(
                screen.getByRole('button', { name: 'AUTO_SHUTDOWN_BANNER_CANCEL' })
            ).toBeInTheDocument();

            act(() => {
                jest.advanceTimersByTime(3000);
            });
            expect(screen.getByRole('alert')).toHaveTextContent('AUTO_SHUTDOWN_BANNER_MESSAGE:39');
        });

        it('cancels through the service and hides the banner', async () => {
            mockedGetStatus.mockResolvedValueOnce(warningStatus).mockResolvedValue(idleStatus);
            mockedCancel.mockResolvedValue(idleStatus);
            renderBanner();

            fireEvent.click(
                await screen.findByRole('button', { name: 'AUTO_SHUTDOWN_BANNER_CANCEL' })
            );

            await waitFor(() => expect(screen.queryByRole('alert')).not.toBeInTheDocument());
            expect(mockedCancel).toHaveBeenCalledTimes(1);
        });

        it('shows the server message when cancelling is refused', async () => {
            mockedGetStatus.mockResolvedValue(warningStatus);
            mockedCancel.mockRejectedValue({ response: { data: { error: 'fora do aviso' } } });
            renderBanner();

            fireEvent.click(
                await screen.findByRole('button', { name: 'AUTO_SHUTDOWN_BANNER_CANCEL' })
            );

            expect(await screen.findByText('fora do aviso')).toBeInTheDocument();
        });
    });
});
