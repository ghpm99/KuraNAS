import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import type { ReactNode } from 'react';
import { render } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import VideoFolderBrowser from './VideoFolderBrowser';
import { buildServerPage } from '@/shared/test/buildServerPage';
import { getVideoLibraryFolders, getVideoLibraryFolderVideos } from '@/service/videoPlayback';

jest.mock('@/features/videos/hooks/useVideoStarToggle/useVideoStarToggle', () => ({
    useVideoStarToggle: () => ({ toggleStar: jest.fn(), isStarTogglePending: false }),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, variables?: Record<string, string>) =>
            variables ? `${key} ${Object.values(variables).join(' ')}` : key,
    }),
}));

jest.mock('@/service/videoPlayback', () => ({
    getVideoLibraryFolders: jest.fn(),
    getVideoLibraryFolderVideos: jest.fn(),
}));

const renderBrowser = (ui: ReactNode, initialEntry = '/') =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <MemoryRouter initialEntries={[initialEntry]}>{ui}</MemoryRouter>
        </QueryClientProvider>
    );

const mockedFolders = getVideoLibraryFolders as jest.Mock;
const mockedFolderVideos = getVideoLibraryFolderVideos as jest.Mock;

const pageOf = (items: unknown[], hasNext = false) => ({
    items,
    pagination: { page: 1, page_size: 48, has_next: hasNext },
});

describe('VideoFolderBrowser', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedFolderVideos.mockResolvedValue(pageOf([]));
    });

    it('lists root folders with count and cover, then drills into one and plays a video', async () => {
        const onPlayVideo = jest.fn();
        mockedFolders.mockImplementation((parentPath: string) =>
            Promise.resolve(
                parentPath === ''
                    ? pageOf([
                          { path: '/Series', name: 'Series', video_count: 3, cover_file_id: 7 },
                      ])
                    : pageOf([])
            )
        );
        mockedFolderVideos.mockResolvedValue(
            pageOf([
                {
                    id: 11,
                    name: 'Ep 1.mkv',
                    path: '/Series/Ep 1.mkv',
                    parent_path: '/Series',
                    format: '.mkv',
                    size: 1,
                },
            ])
        );

        renderBrowser(<VideoFolderBrowser onPlayVideo={onPlayVideo} />);

        const folderCard = await screen.findByRole('button', { name: /Series/ });
        expect(folderCard.querySelector('img')?.getAttribute('src')).toContain(
            '/files/video-thumbnail/7'
        );
        expect(mockedFolderVideos).not.toHaveBeenCalled();

        fireEvent.click(folderCard);

        const videoRow = await screen.findByRole('button', { name: 'VIDEO_PLAY_ARIA Ep 1.mkv' });
        expect(mockedFolders).toHaveBeenCalledWith('/Series', 1, 48);
        expect(mockedFolderVideos).toHaveBeenCalledWith('/Series', 1, 24);

        fireEvent.click(videoRow);
        expect(onPlayVideo).toHaveBeenCalledWith(11, null);
    });

    it('shows the empty state when there are no folders at the roots', async () => {
        mockedFolders.mockResolvedValue(pageOf([]));

        renderBrowser(<VideoFolderBrowser onPlayVideo={jest.fn()} />);

        expect(await screen.findByText('VIDEO_FOLDERS_EMPTY')).toBeInTheDocument();
    });

    it('offers load more when the folders have another page', async () => {
        mockedFolders.mockResolvedValue(
            pageOf([{ path: '/A', name: 'A', video_count: 1, cover_file_id: 0 }], true)
        );

        renderBrowser(<VideoFolderBrowser onPlayVideo={jest.fn()} />);

        await screen.findByRole('button', { name: 'VIDEO_FOLDER_OPEN A' });
        fireEvent.click(await screen.findByRole('button', { name: 'LOAD_MORE' }));
        await waitFor(() => expect(mockedFolders).toHaveBeenCalledWith('', 2, 48));
    });

    it('restores the folder from the url and goes back through the breadcrumb', async () => {
        mockedFolders.mockResolvedValue(pageOf([]));

        renderBrowser(<VideoFolderBrowser onPlayVideo={jest.fn()} />, '/?folder=/Series/S1');

        await waitFor(() => expect(mockedFolderVideos).toHaveBeenCalledWith('/Series/S1', 1, 24));
        fireEvent.click(screen.getByRole('button', { name: 'Series' }));
        await waitFor(() => expect(mockedFolders).toHaveBeenCalledWith('/Series', 1, 48));
        fireEvent.click(screen.getByRole('button', { name: 'VIDEO_SECTION_FOLDERS' }));
        await waitFor(() => expect(mockedFolders).toHaveBeenCalledWith('', 1, 48));
    });

    const buildFolder = (name: string) => ({
        path: `/${name}`,
        name,
        video_count: 1,
        cover_file_id: 0,
    });

    it('keeps the loaded folders when the next page fails and retries only that page', async () => {
        mockedFolders
            .mockResolvedValueOnce(
                buildServerPage([buildFolder('A')], { page: 1, pageSize: 48, hasNext: true })
            )
            .mockRejectedValueOnce(new Error('boom'))
            .mockResolvedValueOnce(
                buildServerPage([buildFolder('B')], { page: 2, pageSize: 48, hasNext: false })
            );

        renderBrowser(<VideoFolderBrowser onPlayVideo={jest.fn()} />);

        await screen.findByRole('button', { name: 'VIDEO_FOLDER_OPEN A' });
        fireEvent.click(await screen.findByRole('button', { name: 'LOAD_MORE' }));

        expect(await screen.findByRole('alert')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'VIDEO_FOLDER_OPEN A' })).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'LOAD_MORE' })).not.toBeInTheDocument();

        fireEvent.click(within(screen.getByRole('alert')).getByRole('button'));

        expect(
            await screen.findByRole('button', { name: 'VIDEO_FOLDER_OPEN B' })
        ).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'VIDEO_FOLDER_OPEN A' })).toBeInTheDocument();
        expect(mockedFolders).toHaveBeenNthCalledWith(3, '', 2, 48);
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });

    it('keeps the loaded videos when the next page fails and retries only that page', async () => {
        const buildVideo = (id: number) => ({
            id,
            name: `Clip ${id}`,
            path: `/S/Clip ${id}.mp4`,
            parent_path: '/S',
            format: '.mp4',
            size: 1,
        });
        mockedFolders.mockResolvedValue(
            buildServerPage([], { page: 1, pageSize: 48, hasNext: false })
        );
        mockedFolderVideos
            .mockResolvedValueOnce(
                buildServerPage([buildVideo(1)], { page: 1, pageSize: 24, hasNext: true })
            )
            .mockRejectedValueOnce(new Error('boom'))
            .mockResolvedValueOnce(
                buildServerPage([buildVideo(2)], { page: 2, pageSize: 24, hasNext: false })
            );

        renderBrowser(<VideoFolderBrowser onPlayVideo={jest.fn()} />, '/?folder=/S');

        await screen.findByRole('button', { name: 'VIDEO_PLAY_ARIA Clip 1' });
        fireEvent.click(await screen.findByRole('button', { name: 'LOAD_MORE' }));

        expect(await screen.findByRole('alert')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'VIDEO_PLAY_ARIA Clip 1' })).toBeInTheDocument();

        fireEvent.click(within(screen.getByRole('alert')).getByRole('button'));

        expect(
            await screen.findByRole('button', { name: 'VIDEO_PLAY_ARIA Clip 2' })
        ).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'VIDEO_PLAY_ARIA Clip 1' })).toBeInTheDocument();
        expect(mockedFolderVideos).toHaveBeenNthCalledWith(3, '/S', 2, 24);
    });

    it('shows the error state, not the empty one, when the server is down from the start', async () => {
        mockedFolders.mockRejectedValue(new Error('down'));

        renderBrowser(<VideoFolderBrowser onPlayVideo={jest.fn()} />);

        expect(await screen.findByRole('alert')).toBeInTheDocument();
        expect(screen.queryByText('VIDEO_FOLDERS_EMPTY')).not.toBeInTheDocument();
    });

    it('keeps the folders visible when only the videos list fails', async () => {
        mockedFolders.mockResolvedValue(
            buildServerPage([buildFolder('A')], { page: 1, pageSize: 48, hasNext: false })
        );
        mockedFolderVideos.mockRejectedValue(new Error('down'));

        renderBrowser(<VideoFolderBrowser onPlayVideo={jest.fn()} />, '/?folder=/S');

        expect(await screen.findByRole('alert')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'VIDEO_FOLDER_OPEN A' })).toBeInTheDocument();
    });
});
