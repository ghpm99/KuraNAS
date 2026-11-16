import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render } from '@testing-library/react';
import { SnackbarProvider } from 'notistack';
import type { ReactNode } from 'react';
import { MemoryRouter } from 'react-router-dom';

export const createVacuumQueryClient = (): QueryClient =>
    new QueryClient({
        defaultOptions: {
            queries: { retry: false },
            mutations: { retry: false },
        },
    });

export const renderWithVacuumProviders = (
    ui: ReactNode,
    queryClient: QueryClient = createVacuumQueryClient()
) => ({
    queryClient,
    ...render(
        <QueryClientProvider client={queryClient}>
            <SnackbarProvider>
                <MemoryRouter>{ui}</MemoryRouter>
            </SnackbarProvider>
        </QueryClientProvider>
    ),
});
