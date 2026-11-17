import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import { MemoryRouter } from 'react-router-dom';
import AnalyticsReindexAction from './AnalyticsReindexAction';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

describe('components/analytics/AnalyticsReindexAction', () => {
    it('renders without any service mock and offers an enabled reindex button', () => {
        const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
        render(
            <QueryClientProvider client={client}>
                <SnackbarProvider>
                    <MemoryRouter>
                        <AnalyticsReindexAction onFinished={jest.fn()} />
                    </MemoryRouter>
                </SnackbarProvider>
            </QueryClientProvider>
        );

        expect(screen.getByRole('button', { name: 'ANALYTICS_ACTION_REINDEX' })).toBeEnabled();
    });
});
