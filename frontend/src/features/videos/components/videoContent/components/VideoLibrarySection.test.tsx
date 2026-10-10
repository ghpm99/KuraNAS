import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render as renderUnwrapped, screen } from '@testing-library/react';
import type { ReactElement } from 'react';
import VideoLibrarySection from './VideoLibrarySection';
import type { VideoFileDto, VideoPlaylistDto } from '@/service/videoPlayback';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string) => {
            const map: Record<string, string> = {
                VIDEO_ALL: 'All Videos',
                VIDEO_ALL_DESC: 'Browse every video',
                VIDEO_SEARCH_PLACEHOLDER: 'Search videos',
                VIDEO_PLAY: 'Play',
                VIDEO_ADD: 'Add',
                VIDEO_ALREADY_ADDED: 'Already Added',
                LOADING: 'Loading',
                LOAD_MORE: 'Load More',
            };
            return map[key] ?? key;
        },
    }),
}));

jest.mock('@/service/apiUrl', () => ({
    getApiV1BaseUrl: () => 'http://localhost:8000/api/v1',
}));

const videos: VideoFileDto[] = [
    {
        id: 5,
        name: 'Episode 5',
        path: '/shows/episode-5.mp4',
        parent_path: '/shows',
        format: 'mp4',
        size: 500,
    },
];

const playlists: VideoPlaylistDto[] = [
    {
        id: 1,
        type: 'series',
        source_path: '/shows',
        name: 'Season 1',
        is_hidden: false,
        is_auto: true,
        group_mode: 'prefix',
        classification: 'series',
        item_count: 1,
        cover_video_id: 5,
        created_at: '2026-03-16T10:00:00Z',
        updated_at: '2026-03-16T10:00:00Z',
        last_played_at: null,
        items: [],
    },
    {
        id: 2,
        type: 'custom',
        source_path: '/custom',
        name: 'Favorites',
        is_hidden: false,
        is_auto: false,
        group_mode: 'single',
        classification: 'personal',
        item_count: 1,
        cover_video_id: null,
        created_at: '2026-03-16T10:00:00Z',
        updated_at: '2026-03-16T10:00:00Z',
        last_played_at: null,
        items: [],
    },
];

const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });

const withQueryClient = (ui: ReactElement) => (
    <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>
);

const render = (ui: ReactElement) => renderUnwrapped(withQueryClient(ui));

describe('videos/videoContent/VideoLibrarySection', () => {
    it('renders without services or optional data', () => {
        render(
            <VideoLibrarySection
                videos={[]}
                playlists={[]}
                search=""
                selectedPlaylistPerVideo={{}}
                isAddingToPlaylist={false}
                isFetchingMoreVideos={false}
                hasMoreVideos={false}
                onSearchChange={jest.fn()}
                onSelectPlaylistForVideo={jest.fn()}
                onPlayVideo={jest.fn()}
                onAddVideo={jest.fn()}
                onLoadMore={jest.fn()}
            />
        );

        expect(screen.getByText('All Videos')).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'Load More' })).not.toBeInTheDocument();
    });

    it('uses the first playlist as fallback selection and forwards user actions', () => {
        const onSearchChange = jest.fn();
        const onSelectPlaylistForVideo = jest.fn();
        const onPlayVideo = jest.fn();
        const onAddVideo = jest.fn();
        const onLoadMore = jest.fn();

        render(
            <VideoLibrarySection
                videos={videos}
                playlists={playlists}
                search=""
                selectedPlaylistPerVideo={{}}
                isAddingToPlaylist={false}
                isFetchingMoreVideos={false}
                hasMoreVideos={false}
                onSearchChange={onSearchChange}
                onSelectPlaylistForVideo={onSelectPlaylistForVideo}
                onPlayVideo={onPlayVideo}
                onAddVideo={onAddVideo}
                onLoadMore={onLoadMore}
            />
        );

        expect(screen.getByText('All Videos')).toBeInTheDocument();
        expect(screen.getByText('Browse every video')).toBeInTheDocument();
        expect(screen.getByRole('img', { name: 'Episode 5' })).toHaveAttribute(
            'src',
            'http://localhost:8000/api/v1/files/video-thumbnail/5?width=240&height=135'
        );
        expect(screen.getByRole('combobox')).toHaveValue('1');

        fireEvent.change(screen.getByPlaceholderText('Search videos'), {
            target: { value: 'epi' },
        });
        fireEvent.change(screen.getByRole('combobox'), { target: { value: '2' } });
        fireEvent.click(screen.getByRole('button', { name: 'Play' }));
        fireEvent.click(screen.getByRole('button', { name: 'Add' }));

        expect(onSearchChange).toHaveBeenCalledWith('epi');
        expect(onSelectPlaylistForVideo).toHaveBeenCalledWith(5, 2);
        expect(onPlayVideo).toHaveBeenCalledWith(5, null);
        expect(onAddVideo).toHaveBeenCalledWith(5);
        expect(onLoadMore).not.toHaveBeenCalled();
        expect(screen.queryByRole('button', { name: 'Load More' })).not.toBeInTheDocument();
    });

    it('handles the footer loading states', () => {
        const onLoadMore = jest.fn();
        const { rerender } = render(
            <VideoLibrarySection
                videos={videos}
                playlists={playlists}
                search=""
                selectedPlaylistPerVideo={{ 5: 2 }}
                isAddingToPlaylist={false}
                isFetchingMoreVideos={true}
                hasMoreVideos
                onSearchChange={jest.fn()}
                onSelectPlaylistForVideo={jest.fn()}
                onPlayVideo={jest.fn()}
                onAddVideo={jest.fn()}
                onLoadMore={onLoadMore}
            />
        );

        expect(screen.getByRole('button', { name: 'Load More' })).toBeDisabled();

        rerender(
            withQueryClient(
                <VideoLibrarySection
                    videos={videos}
                    playlists={playlists}
                    search=""
                    selectedPlaylistPerVideo={{ 5: 2 }}
                    isAddingToPlaylist={false}
                    isFetchingMoreVideos={false}
                    hasMoreVideos
                    onSearchChange={jest.fn()}
                    onSelectPlaylistForVideo={jest.fn()}
                    onPlayVideo={jest.fn()}
                    onAddVideo={jest.fn()}
                    onLoadMore={onLoadMore}
                />
            )
        );

        fireEvent.click(screen.getByRole('button', { name: 'Load More' }));
        expect(onLoadMore).toHaveBeenCalled();
    });

    it('marks a library video as watched', () => {
        const onSetWatched = jest.fn();
        render(
            <VideoLibrarySection
                videos={videos}
                playlists={playlists}
                search=""
                selectedPlaylistPerVideo={{}}
                isAddingToPlaylist={false}
                isFetchingMoreVideos={false}
                hasMoreVideos={false}
                onSearchChange={jest.fn()}
                onSelectPlaylistForVideo={jest.fn()}
                onPlayVideo={jest.fn()}
                onAddVideo={jest.fn()}
                onLoadMore={jest.fn()}
                onSetWatched={onSetWatched}
            />
        );

        fireEvent.click(screen.getByRole('button', { name: 'VIDEO_MARK_WATCHED' }));

        expect(onSetWatched).toHaveBeenCalledWith(5, true);
    });
});
