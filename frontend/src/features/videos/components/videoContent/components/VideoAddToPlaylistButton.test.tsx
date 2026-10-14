import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen } from '@testing-library/react';
import VideoAddToPlaylistButton from './VideoAddToPlaylistButton';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock };

const renderButton = (onAddVideo = jest.fn(), isAddingToPlaylist = false) => {
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <VideoAddToPlaylistButton
                videoId={5}
                isAddingToPlaylist={isAddingToPlaylist}
                onAddVideo={onAddVideo}
            />
        </QueryClientProvider>
    );
    return onAddVideo;
};

describe('features/videos/VideoAddToPlaylistButton', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockImplementation((url: string) =>
            Promise.resolve({
                data: url.includes('by-video')
                    ? []
                    : {
                          items: [{ id: 7, name: 'Season 1', type: 'custom' }],
                          pagination: { page: 1, page_size: 30, has_next: false, has_prev: false },
                      },
            })
        );
    });

    it('does not ask the backend nor add anything until the menu is opened', () => {
        const onAddVideo = renderButton();

        expect(mockedApi.get).not.toHaveBeenCalled();
        expect(onAddVideo).not.toHaveBeenCalled();
    });

    it('opens the menu and adds the video to the explicitly chosen playlist', async () => {
        const onAddVideo = renderButton();

        fireEvent.click(screen.getByRole('button', { name: 'VIDEO_ADD_TO_PLAYLIST' }));
        fireEvent.click(await screen.findByText('Season 1'));

        expect(onAddVideo).toHaveBeenCalledWith(5, 7);
    });

    it('blocks opening the menu while an add is in flight', () => {
        renderButton(jest.fn(), true);

        expect(screen.getByRole('button', { name: 'VIDEO_ADD_TO_PLAYLIST' })).toBeDisabled();
    });

    it('creates a playlist already containing the video from the menu', async () => {
        const onCreatePlaylistWithVideo = jest.fn();
        render(
            <QueryClientProvider
                client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
            >
                <VideoAddToPlaylistButton
                    videoId={5}
                    isAddingToPlaylist={false}
                    onAddVideo={jest.fn()}
                    onCreatePlaylistWithVideo={onCreatePlaylistWithVideo}
                />
            </QueryClientProvider>
        );

        fireEvent.click(screen.getByRole('button', { name: 'VIDEO_ADD_TO_PLAYLIST' }));
        fireEvent.click(await screen.findByText('VIDEO_NEW_PLAYLIST'));
        fireEvent.change(await screen.findByLabelText('VIDEO_NEW_PLAYLIST_NAME'), {
            target: { value: 'Favoritos' },
        });
        fireEvent.click(screen.getByRole('button', { name: 'ACTION_CREATE_ADD' }));

        expect(onCreatePlaylistWithVideo).toHaveBeenCalledWith('Favoritos', 5);
    });

    it('does not offer playlist creation when no creation handler is wired', async () => {
        renderButton();

        fireEvent.click(screen.getByRole('button', { name: 'VIDEO_ADD_TO_PLAYLIST' }));
        await screen.findByText('Season 1');

        expect(screen.queryByText('VIDEO_NEW_PLAYLIST')).not.toBeInTheDocument();
    });
});
