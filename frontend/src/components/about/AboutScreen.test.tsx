import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import AboutScreen from './AboutScreen';
import {
    AboutContextProvider,
    type AboutContextType,
} from '@/components/providers/aboutProvider/AboutContext';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string) => key,
    }),
}));

jest.mock('@/service/update', () => ({
    getUpdateStatus: jest.fn().mockResolvedValue({
        current_version: '2.4.0',
        latest_version: '2.4.0',
        update_available: false,
        release_url: '',
        release_date: '',
        release_notes: '',
        asset_name: '',
        asset_size: 0,
    }),
    applyUpdate: jest.fn(),
}));

const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
});

const aboutValue: AboutContextType = {
    version: '2.4.0',
    commit_hash: 'abc123def456',
    platform: 'linux',
    path: '/srv/media',
    lang: 'pt-BR',
    enable_workers: true,
    uptime: '1h 12m',
    statup_time: '2026-03-15T12:00:00.000Z',
    gin_mode: 'release',
    gin_version: '1.10.0',
    go_version: 'go1.24.0',
    node_version: 'v24.1.0',
};

describe('components/about/AboutScreen', () => {
    beforeEach(() => {
        jest.useFakeTimers();
        Object.assign(navigator, {
            clipboard: {
                writeText: jest.fn().mockResolvedValue(undefined),
            },
        });
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    const renderAbout = () =>
        render(
            <QueryClientProvider client={queryClient}>
                <SnackbarProvider>
                    <MemoryRouter>
                        <AboutContextProvider value={aboutValue}>
                            <AboutScreen />
                        </AboutContextProvider>
                    </MemoryRouter>
                </SnackbarProvider>
            </QueryClientProvider>
        );

    it('renders runtime, build details and technical tools', () => {
        renderAbout();

        expect(screen.getByRole('heading', { name: 'ABOUT_PAGE_TITLE' })).toBeInTheDocument();
        expect(screen.getByText('2.4.0')).toBeInTheDocument();
        expect(screen.getByText('/srv/media')).toBeInTheDocument();
        expect(screen.getByText('abc123def456')).toBeInTheDocument();
        expect(screen.getByText('ABOUT_TOOL_ANALYTICS_TITLE')).toBeInTheDocument();
        expect(screen.getByText('ABOUT_TOOL_SETTINGS_TITLE')).toBeInTheDocument();

        const links = screen.getAllByRole('link', { name: 'ABOUT_OPEN_TOOL' });
        expect(links).toHaveLength(2);
        expect(links[0]).toHaveAttribute('href', '/analytics');
        expect(links[1]).toHaveAttribute('href', '/settings');
    });

    it('copies the commit hash and shows feedback', async () => {
        renderAbout();

        await act(async () => {
            fireEvent.click(screen.getByRole('button', { name: 'ABOUT_COPY_COMMIT' }));
        });

        await waitFor(() => {
            expect(navigator.clipboard.writeText).toHaveBeenCalledWith('abc123def456');
        });
        await waitFor(() => {
            expect(screen.getByRole('button', { name: 'ABOUT_COMMIT_COPIED' })).toBeInTheDocument();
        });

        await act(async () => {
            jest.advanceTimersByTime(2000);
        });

        await waitFor(() => {
            expect(screen.getByRole('button', { name: 'ABOUT_COPY_COMMIT' })).toBeInTheDocument();
        });
    });
});
