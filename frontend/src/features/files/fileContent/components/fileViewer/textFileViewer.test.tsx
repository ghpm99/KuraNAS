import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactElement } from 'react';
import TextFileViewer from './textFileViewer';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, options?: Record<string, string>) =>
            options?.size ? `${key}:${options.size}` : key,
    }),
}));

const renderWithQuery = (ui: ReactElement) =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            {ui}
        </QueryClientProvider>
    );

describe('TextFileViewer without service mocks', () => {
    it('shows the loading state and then the error alert when the backend is absent', async () => {
        renderWithQuery(<TextFileViewer file={{ id: 1, size: 10 }} />);

        expect(screen.getByText('TEXT_VIEWER_LOADING')).toBeInTheDocument();
        expect(await screen.findByText('TEXT_VIEWER_ERROR')).toBeInTheDocument();
    });
});
