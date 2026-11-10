import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import TrashSummaryLine from './TrashSummaryLine';
import { apiBase } from '@/service';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';

const renderSummary = () =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <TrashSummaryLine />
        </QueryClientProvider>
    );

describe('TrashSummaryLine', () => {
    afterEach(() => {
        jest.restoreAllMocks();
    });

    it('renders nothing without backend', async () => {
        const { container } = renderWithoutBackend(<TrashSummaryLine />);

        await waitFor(() => expect(container.querySelector('p')).toBeNull());
    });

    it('reads the totals from GET /trash/summary', async () => {
        const getSummary = jest
            .spyOn(apiBase, 'get')
            .mockResolvedValue({ data: { items: 3, size_bytes: 2048 } });
        renderSummary();

        expect(await screen.findByText('TRASH_SUMMARY')).toBeInTheDocument();
        expect(getSummary).toHaveBeenCalledWith('/trash/summary');
    });

    it('uses the singular text for one item', async () => {
        jest.spyOn(apiBase, 'get').mockResolvedValue({ data: { items: 1, size_bytes: 10 } });
        renderSummary();

        expect(await screen.findByText('TRASH_SUMMARY_ONE')).toBeInTheDocument();
    });

    it('stays hidden for an empty trash and for a payload without totals', async () => {
        const getSummary = jest.spyOn(apiBase, 'get').mockResolvedValue({ data: { items: 0 } });
        const { container, unmount } = renderSummary();
        await waitFor(() => expect(getSummary).toHaveBeenCalled());
        expect(container.querySelector('p')).toBeNull();
        unmount();

        getSummary.mockResolvedValue({ data: {} });
        const { container: partialContainer } = renderSummary();
        await waitFor(() => expect(getSummary).toHaveBeenCalledTimes(2));
        expect(partialContainer.querySelector('p')).toBeNull();
    });
});
