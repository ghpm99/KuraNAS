import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import TextFileViewer from './textFileViewer';

const mockGet = jest.fn();

jest.mock('@/service', () => ({
    apiBase: { get: (...args: unknown[]) => mockGet(...args) },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, options?: Record<string, string>) =>
            options?.size ? `${key}:${options.size}` : key,
    }),
}));

const renderViewer = (fileSize: number, body: string) => {
    mockGet.mockResolvedValue({ data: body });
    return render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <TextFileViewer file={{ id: 9, size: fileSize }} />
        </QueryClientProvider>
    );
};

describe('TextFileViewer seam (apiBase only)', () => {
    beforeEach(() => mockGet.mockReset());

    it('requests at most the first 512 KB through an HTTP Range header', async () => {
        renderViewer(100, 'hello world');

        expect(await screen.findByText('hello world')).toBeInTheDocument();
        expect(mockGet).toHaveBeenCalledTimes(1);
        expect(mockGet).toHaveBeenCalledWith(
            '/files/blob/9',
            expect.objectContaining({
                headers: { Range: 'bytes=0-524287' },
                responseType: 'text',
            })
        );
        expect(screen.queryByText(/TEXT_VIEWER_TRUNCATED/)).toBeNull();
    });

    it('warns that the file was truncated when it exceeds the preview limit', async () => {
        renderViewer(600 * 1024, 'partial content');

        expect(await screen.findByText('partial content')).toBeInTheDocument();
        expect(screen.getByText(/TEXT_VIEWER_TRUNCATED:/)).toBeInTheDocument();
    });

    it('cuts the body when the server ignores the Range header', async () => {
        renderViewer(100, 'x'.repeat(600 * 1024));

        await waitFor(() => expect(screen.getByText(/TEXT_VIEWER_TRUNCATED:/)).toBeInTheDocument());
        const renderedText = screen.getByText(/^x+$/).textContent ?? '';
        expect(renderedText.length).toBe(512 * 1024);
    });

    it('shows an empty notice for an empty file', async () => {
        renderViewer(0, '');

        expect(await screen.findByText('TEXT_VIEWER_EMPTY')).toBeInTheDocument();
    });

    it('shows the error alert when the request fails', async () => {
        mockGet.mockRejectedValue(new Error('boom'));
        render(
            <QueryClientProvider
                client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
            >
                <TextFileViewer file={{ id: 9, size: 5 }} />
            </QueryClientProvider>
        );

        expect(await screen.findByText('TEXT_VIEWER_ERROR')).toBeInTheDocument();
    });

    it('toggles line wrapping on the monospace block', async () => {
        renderViewer(10, 'some long line');

        const block = (await screen.findByText('some long line')) as HTMLElement;
        expect(getComputedStyle(block).whiteSpace).toBe('pre');

        fireEvent.click(screen.getByRole('switch', { name: 'TEXT_VIEWER_WRAP' }));

        expect(getComputedStyle(block).whiteSpace).toBe('pre-wrap');
    });
});
