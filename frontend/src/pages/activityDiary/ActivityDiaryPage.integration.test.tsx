import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import ActivityDiaryPage from '.';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn() },
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock };

const buildEntry = (id: number) => ({
    id,
    name: `Atividade ${id}`,
    description: '',
    start_time: '2026-06-01T10:00:00Z',
    end_time: { HasValue: true, Value: '2026-06-01T11:00:00Z' },
    duration: 3600,
});

const renderPage = () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <SnackbarProvider>
                <ActivityDiaryPage />
            </SnackbarProvider>
        </QueryClientProvider>
    );
};

describe('pages/activityDiary (page layout and pagination)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('renders the single h1 and an error state when the backend is absent', async () => {
        mockedApi.get.mockRejectedValue({ response: { data: { error: 'servidor fora do ar' } } });

        renderPage();

        expect(screen.getAllByRole('heading', { level: 1 })).toHaveLength(1);
        expect(await screen.findByText('servidor fora do ar')).toBeInTheDocument();
        expect(screen.getByRole('alert')).toBeInTheDocument();
    });

    it('requests the first page with page_size and loads the next page on demand', async () => {
        mockedApi.get.mockImplementation((url: string, config?: { params?: { page: number } }) => {
            if (url === '/diary/summary') return Promise.resolve({ data: { total_activities: 0 } });
            const page = config?.params?.page ?? 1;
            return Promise.resolve({
                data: {
                    items: [buildEntry(page)],
                    pagination: { page, page_size: 20, has_next: page === 1, has_prev: page > 1 },
                },
            });
        });

        renderPage();
        await screen.findByText('Atividade 1');
        expect(mockedApi.get).toHaveBeenCalledWith('/diary/', {
            params: { page: 1, page_size: 20 },
        });

        fireEvent.click(screen.getByRole('button', { name: 'LOAD_MORE' }));

        await screen.findByText('Atividade 2');
        expect(mockedApi.get).toHaveBeenCalledWith('/diary/', {
            params: { page: 2, page_size: 20 },
        });
        await waitFor(() => expect(screen.queryByRole('button', { name: 'LOAD_MORE' })).toBeNull());
    });
});
