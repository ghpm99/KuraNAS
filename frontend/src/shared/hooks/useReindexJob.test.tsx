import { ReactNode } from 'react';
import { renderHook } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import useReindexJob from './useReindexJob';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const createWrapper = () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={queryClient}>
            <SnackbarProvider>{children}</SnackbarProvider>
        </QueryClientProvider>
    );
};

describe('useReindexJob', () => {
    it('renders idle without a backend', () => {
        const { result } = renderHook(() => useReindexJob(), { wrapper: createWrapper() });

        expect(result.current.isRunning).toBe(false);
        expect(result.current.job).toBeNull();
        expect(result.current.progress).toBe(0);
        expect(result.current.errorMessage).toBeNull();
    });
});
