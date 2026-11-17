import { act, render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import AutoShutdownBanner from './AutoShutdownBanner';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

describe('layout/AutoShutdownBanner without a service mock', () => {
    it('renders nothing when the server is absent', async () => {
        const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
        render(
            <QueryClientProvider client={client}>
                <AutoShutdownBanner />
            </QueryClientProvider>
        );

        await act(async () => {
            await new Promise((resolve) => setTimeout(resolve, 50));
        });
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });
});
