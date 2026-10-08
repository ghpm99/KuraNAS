import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import type { ReactNode } from 'react';
import DiskLocationRow from './diskLocationRow';

const withProviders = (children: ReactNode) => (
    <QueryClientProvider
        client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
        <SnackbarProvider>{children}</SnackbarProvider>
    </QueryClientProvider>
);

describe('DiskLocationRow without service mocks', () => {
    it('renders nothing and does not crash while the backend is absent', async () => {
        const { container } = render(withProviders(<DiskLocationRow fileId={1} />));

        expect(container.querySelector('code')).toBeNull();
        expect(screen.queryByText('FILE_DISK_LOCATION')).toBeNull();
    });

    it('renders nothing without a file id', () => {
        render(withProviders(<DiskLocationRow />));

        expect(screen.queryByText('FILE_DISK_LOCATION')).toBeNull();
    });
});
