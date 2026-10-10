import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import VideoMovieBrowser from './VideoMovieBrowser';
import { getVideoLibraryMovies } from '@/service/videoPlayback';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, variables?: Record<string, string>) =>
            variables ? `${key} ${Object.values(variables).join(' ')}` : key,
    }),
}));

jest.mock('@/service/videoPlayback', () => ({
    getVideoLibraryMovies: jest.fn(),
}));

const mockedMovies = getVideoLibraryMovies as jest.Mock;

const renderBrowser = (onPlayVideo = jest.fn()) =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <VideoMovieBrowser onPlayVideo={onPlayVideo} />
        </QueryClientProvider>
    );

const movie = (id: number, name: string) => ({
    id,
    name,
    path: `/${name}`,
    parent_path: '/',
    format: '.mkv',
    size: 1,
});

const pageOf = (items: unknown[], hasNext = false) => ({
    items,
    pagination: { page: 1, page_size: 24, has_next: hasNext },
});

describe('VideoMovieBrowser', () => {
    beforeEach(() => jest.clearAllMocks());

    it('lists the movies with a thumbnail and plays one on click', async () => {
        const onPlayVideo = jest.fn();
        mockedMovies.mockResolvedValue(pageOf([movie(5, 'Alien.mkv')]));

        renderBrowser(onPlayVideo);

        const card = await screen.findByRole('button', { name: /Alien.mkv/ });
        expect(card.querySelector('img')?.getAttribute('src')).toContain(
            '/files/video-thumbnail/5'
        );
        expect(mockedMovies).toHaveBeenCalledWith('name', 1, 24);

        fireEvent.click(card);
        expect(onPlayVideo).toHaveBeenCalledWith(5, null);
    });

    it('requests the next page through load more', async () => {
        mockedMovies.mockResolvedValue(pageOf([movie(1, 'A.mkv')], true));

        renderBrowser();

        fireEvent.click(await screen.findByRole('button', { name: 'LOAD_MORE' }));
        await waitFor(() => expect(mockedMovies).toHaveBeenCalledWith('name', 2, 24));
    });

    it('refetches from the first page when the sort changes', async () => {
        mockedMovies.mockResolvedValue(pageOf([movie(1, 'A.mkv')]));

        renderBrowser();

        await screen.findByRole('button', { name: /A.mkv/ });
        fireEvent.mouseDown(screen.getByRole('combobox'));
        fireEvent.click(await screen.findByRole('option', { name: 'VIDEO_MOVIES_SORT_RECENT' }));

        await waitFor(() => expect(mockedMovies).toHaveBeenCalledWith('recent', 1, 24));
    });

    it('shows the empty state when there are no movies', async () => {
        mockedMovies.mockResolvedValue(pageOf([]));

        renderBrowser();

        expect(await screen.findByText('VIDEO_SECTION_MOVIES_EMPTY')).toBeInTheDocument();
    });

    it('shows the error state and retries', async () => {
        mockedMovies.mockRejectedValueOnce(new Error('down'));
        mockedMovies.mockResolvedValue(pageOf([movie(2, 'B.mkv')]));

        renderBrowser();

        fireEvent.click(await screen.findByRole('button', { name: 'TRY_AGAIN' }));
        expect(await screen.findByRole('button', { name: /B.mkv/ })).toBeInTheDocument();
    });
});
