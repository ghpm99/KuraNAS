import { render, type RenderResult } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import type { ReactNode } from 'react';
import { MemoryRouter } from 'react-router-dom';
import I18nProvider from '@/components/i18n/provider';

// renderWithoutBackend mounts a component inside the REAL app-level context it
// structurally needs (query cache, i18n, snackbars, router) but with NO service
// or backend mock. Every API call simply fails against a non-existent server.
//
// This is the harness for the resilience-first rule (frontend/CLAUDE.md): the
// first test of every component renders it through here and asserts it does not
// throw. A component that assumes a complete/!nil backend response crashes here,
// which is exactly what we want to catch.
export const renderWithoutBackend = (ui: ReactNode): RenderResult => {
    const queryClient = new QueryClient({
        defaultOptions: {
            queries: { retry: false },
            mutations: { retry: false },
        },
    });

    return render(
        <QueryClientProvider client={queryClient}>
            <I18nProvider>
                <SnackbarProvider>
                    <MemoryRouter>{ui}</MemoryRouter>
                </SnackbarProvider>
            </I18nProvider>
        </QueryClientProvider>
    );
};

// expectRendersWithoutBackend is the one-liner the smoke tests use:
//   it('renders without backend', () => expectRendersWithoutBackend(<X />));
export const expectRendersWithoutBackend = (ui: ReactNode): void => {
    expect(() => renderWithoutBackend(ui)).not.toThrow();
};
