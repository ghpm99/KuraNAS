import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import AutoShutdownBanner from './AutoShutdownBanner';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock };

const idleStatus = {
    state: 'idle',
    shutdown_at: null,
    seconds_remaining: null,
    postponed_reason: '',
};
const warningStatus = {
    state: 'warning',
    shutdown_at: '2026-11-17T03:00:00Z',
    seconds_remaining: 30,
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

describe('layout/AutoShutdownBanner (seam)', () => {
    beforeEach(() => {
        jest.resetAllMocks();
    });

    it('shows the banner on warning and cancel posts to /auto-shutdown/cancel', async () => {
        let currentStatus: typeof warningStatus | typeof idleStatus = warningStatus;
        mockedApi.get.mockImplementation((url: string) =>
            url === '/auto-shutdown/status'
                ? Promise.resolve({ data: currentStatus })
                : Promise.reject(new Error(`unexpected GET ${url}`))
        );
        mockedApi.post.mockImplementation(() => {
            currentStatus = idleStatus;
            return Promise.resolve({ data: idleStatus });
        });
        renderBanner();

        fireEvent.click(await screen.findByRole('button', { name: 'AUTO_SHUTDOWN_BANNER_CANCEL' }));

        await waitFor(() => expect(screen.queryByRole('alert')).not.toBeInTheDocument());
        expect(mockedApi.post).toHaveBeenCalledTimes(1);
        expect(mockedApi.post).toHaveBeenCalledWith('/auto-shutdown/cancel');
    });

    it('shows the 409 message from the server verbatim', async () => {
        mockedApi.get.mockResolvedValue({ data: warningStatus });
        mockedApi.post.mockRejectedValue({
            response: { status: 409, data: { error: 'nenhum desligamento em aviso' } },
        });
        renderBanner();

        fireEvent.click(await screen.findByRole('button', { name: 'AUTO_SHUTDOWN_BANNER_CANCEL' }));

        expect(await screen.findByText('nenhum desligamento em aviso')).toBeInTheDocument();
    });
});
