import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import DownloadsScreen from './DownloadsScreen';
import { isBrowserExtension } from './useDownloadsScreen';
import { getDownloads } from '@/service/downloads';
import type { DownloadItem } from '@/types/downloads';

jest.mock('@/service/downloads', () => ({
    getDownloads: jest.fn(),
    buildDownloadHref: (item: DownloadItem) => `/api/v1/downloads/${item.id}`,
}));

const mockedGetDownloads = getDownloads as jest.Mock;

const androidItem: DownloadItem = {
    id: 'android',
    name: 'Android App',
    description: 'Native app',
    platform: 'android',
    version: '1.0.0',
    min_os: 'Android 13',
    size_bytes: 2 * 1024 * 1024,
    sha256: 'abc',
    download_url: '/api/v1/downloads/android',
};

const pluginItem: DownloadItem = {
    id: 'plugin',
    name: 'Browser Extension',
    description: '',
    platform: 'browser',
    version: '',
    min_os: '',
    size_bytes: 0,
    sha256: 'def',
    download_url: '/api/v1/downloads/plugin',
};

const renderScreen = () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <MemoryRouter>
                <DownloadsScreen />
            </MemoryRouter>
        </QueryClientProvider>
    );
};

describe('DownloadsScreen', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('shows the loading state while fetching', () => {
        mockedGetDownloads.mockReturnValue(new Promise(() => {}));
        renderScreen();
        expect(screen.getByText('DOWNLOADS_LOADING')).toBeInTheDocument();
    });

    it('shows the error state when the request fails', async () => {
        mockedGetDownloads.mockRejectedValue(new Error('boom'));
        renderScreen();
        await waitFor(() => expect(screen.getByText('DOWNLOADS_ERROR')).toBeInTheDocument());
    });

    it('shows the empty state when there are no apps', async () => {
        mockedGetDownloads.mockResolvedValue([]);
        renderScreen();
        await waitFor(() => expect(screen.getByText('DOWNLOADS_EMPTY')).toBeInTheDocument());
    });

    it('renders app cards with a download link and the plugin instructions', async () => {
        mockedGetDownloads.mockResolvedValue([androidItem, pluginItem]);
        renderScreen();

        await waitFor(() => expect(screen.getByText('Android App')).toBeInTheDocument());
        expect(screen.getByText('Browser Extension')).toBeInTheDocument();

        const link = screen.getAllByRole('link', { name: /DOWNLOADS_BUTTON/ })[0];
        expect(link).toHaveAttribute('href', '/api/v1/downloads/android');

        // a browser extension is present, so the manual install steps render
        expect(screen.getByText('DOWNLOADS_PLUGIN_INSTRUCTIONS_TITLE')).toBeInTheDocument();
    });

    it('hides the plugin instructions when no browser extension is offered', async () => {
        mockedGetDownloads.mockResolvedValue([androidItem]);
        renderScreen();

        await waitFor(() => expect(screen.getByText('Android App')).toBeInTheDocument());
        expect(screen.queryByText('DOWNLOADS_PLUGIN_INSTRUCTIONS_TITLE')).not.toBeInTheDocument();
    });
});

describe('useDownloadsScreen helpers', () => {
    it('detects browser extensions by platform', () => {
        expect(isBrowserExtension(pluginItem)).toBe(true);
        expect(isBrowserExtension(androidItem)).toBe(false);
    });
});
