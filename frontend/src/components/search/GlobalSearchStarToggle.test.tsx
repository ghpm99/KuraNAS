import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import GlobalSearchProvider from './GlobalSearchProvider';
import { apiBase } from '@/service';
import { setFileStarred } from '@/service/files';

const mockEnqueueSnackbar = jest.fn();

jest.mock('notistack', () => ({
    useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn() },
}));

jest.mock('@/service/files', () => ({
    setFileStarred: jest.fn(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) =>
            params ? `${key}:${JSON.stringify(params)}` : key,
    }),
}));

const mockedGet = (apiBase as unknown as { get: jest.Mock }).get;
const mockedSetFileStarred = setFileStarred as jest.Mock;

const searchResponse = {
    query: 'trip',
    files: [
        {
            id: 1,
            name: 'trip.jpg',
            path: '/trip.jpg',
            parent_path: '/',
            format: '.jpg',
            starred: false,
        },
    ],
    folders: [],
    artists: [],
    albums: [],
    playlists: [],
    videos: [],
    images: [],
};

const renderProvider = () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <MemoryRouter>
                <GlobalSearchProvider>
                    <div />
                </GlobalSearchProvider>
            </MemoryRouter>
        </QueryClientProvider>
    );
};

const openSearchWithResult = async () => {
    fireEvent.keyDown(window, { key: 'k', ctrlKey: true });
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'trip' } });
    return screen.findByRole('button', { name: /FILES_FAVORITE_ITEM/ }, { timeout: 3000 });
};

describe('GlobalSearchProvider star toggle', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedGet.mockResolvedValue({ data: searchResponse });
    });

    it('renders the dialog without crashing when the backend is absent', async () => {
        mockedGet.mockRejectedValue(new Error('offline'));
        renderProvider();
        fireEvent.keyDown(window, { key: 'k', ctrlKey: true });
        fireEvent.change(screen.getByRole('combobox'), { target: { value: 'trip' } });
        expect(await screen.findByText('GLOBAL_SEARCH_ERROR_TITLE')).toBeInTheDocument();
    });

    it('flips aria-pressed immediately, calls setFileStarred and keeps the dialog open', async () => {
        mockedSetFileStarred.mockResolvedValue(undefined);
        renderProvider();
        const starButton = await openSearchWithResult();

        fireEvent.click(starButton);

        await waitFor(() =>
            expect(
                screen.getByRole('button', { name: /FILES_UNFAVORITE_ITEM/ })
            ).toHaveAttribute('aria-pressed', 'true')
        );
        expect(mockedSetFileStarred).toHaveBeenCalledWith(1, true);
        expect(screen.getByRole('combobox')).toBeInTheDocument();
    });

    it('keeps opening the active result with Enter', async () => {
        renderProvider();
        await openSearchWithResult();

        fireEvent.keyDown(screen.getByRole('combobox'), { key: 'Enter' });

        await waitFor(() => expect(screen.queryByRole('combobox')).not.toBeInTheDocument());
    });

    it('rolls back and shows the server message when the request fails', async () => {
        mockedSetFileStarred.mockRejectedValue({
            isAxiosError: true,
            response: { data: { error: 'mensagem do servidor' } },
        });
        renderProvider();
        const starButton = await openSearchWithResult();

        fireEvent.click(starButton);

        await waitFor(() =>
            expect(mockEnqueueSnackbar).toHaveBeenCalledWith('mensagem do servidor', {
                variant: 'error',
            })
        );
        expect(screen.getByRole('button', { name: /FILES_FAVORITE_ITEM/ })).toHaveAttribute(
            'aria-pressed',
            'false'
        );
    });
});
