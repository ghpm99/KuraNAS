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

    it('reports the chosen sort key and toggles the direction', () => {
        const onSortChange = jest.fn();
        render(
            <VideoLibrarySection
                videos={videos}
                playlists={playlists}
                search=""
                sort={{ key: 'name', order: 'asc' }}
                selectedPlaylistPerVideo={{}}
                isAddingToPlaylist={false}
                isFetchingMoreVideos={false}
                hasMoreVideos={false}
                onSearchChange={jest.fn()}
                onSortChange={onSortChange}
                onSelectPlaylistForVideo={jest.fn()}
                onPlayVideo={jest.fn()}
                onAddVideo={jest.fn()}
                onLoadMore={jest.fn()}
            />
        );

        expect(screen.getByRole('combobox', { name: 'VIDEO_SORT_LABEL' })).toHaveValue('name');
        fireEvent.change(screen.getByRole('combobox', { name: 'VIDEO_SORT_LABEL' }), {
            target: { value: 'size' },
        });
        fireEvent.click(screen.getByRole('button', { name: 'VIDEO_SORT_ORDER_ASCENDING' }));

        expect(onSortChange).toHaveBeenNthCalledWith(1, { key: 'size', order: 'asc' });
        expect(onSortChange).toHaveBeenNthCalledWith(2, { key: 'name', order: 'desc' });
    });

    it('shows the descending toggle by default and tolerates a missing sort handler', () => {
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

        fireEvent.click(screen.getByRole('button', { name: 'VIDEO_SORT_ORDER_DESCENDING' }));
        fireEvent.change(screen.getByRole('combobox', { name: 'VIDEO_SORT_LABEL' }), {
            target: { value: 'name' },
        });

        expect(screen.getByRole('combobox', { name: 'VIDEO_SORT_LABEL' })).toHaveValue('recent');
    });

    it('shows the filtered empty state with a clear action when filters hide everything', () => {
        const onFiltersChange = jest.fn();
        render(
            <VideoLibrarySection
                videos={[]}
                playlists={[]}
                search=""
                filters={{ formats: ['.mkv'] }}
                selectedPlaylistPerVideo={{}}
                isAddingToPlaylist={false}
                isFetchingMoreVideos={false}
                hasMoreVideos={false}
                onSearchChange={jest.fn()}
                onFiltersChange={onFiltersChange}
                onSelectPlaylistForVideo={jest.fn()}
                onPlayVideo={jest.fn()}
                onAddVideo={jest.fn()}
                onLoadMore={jest.fn()}
            />
        );

        expect(screen.getByText('VIDEO_FILTER_EMPTY')).toBeInTheDocument();
        const [clearFromBar, clearFromEmptyState] = screen.getAllByRole('button', {
            name: 'VIDEO_FILTER_CLEAR',
        });
        fireEvent.click(clearFromEmptyState!);
        fireEvent.click(clearFromBar!);

        expect(onFiltersChange).toHaveBeenCalledTimes(2);
        expect(onFiltersChange).toHaveBeenCalledWith({ formats: [] });
    });

    it('does not show the filtered empty state without active filters', () => {
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

        expect(screen.queryByText('VIDEO_FILTER_EMPTY')).not.toBeInTheDocument();
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
        expect(screen.getAllByRole('combobox')[3]!).toHaveValue('1');

        fireEvent.change(screen.getByPlaceholderText('Search videos'), {
            target: { value: 'epi' },
        });
        fireEvent.change(screen.getAllByRole('combobox')[3]!, { target: { value: '2' } });
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

    it('shows badges for videos with technical fields and omits them otherwise', () => {
        const props = {
            playlists: [],
            search: '',
            selectedPlaylistPerVideo: {},
            isAddingToPlaylist: false,
            isFetchingMoreVideos: false,
            hasMoreVideos: false,
            onSearchChange: jest.fn(),
            onSelectPlaylistForVideo: jest.fn(),
            onPlayVideo: jest.fn(),
            onAddVideo: jest.fn(),
            onLoadMore: jest.fn(),
        };
        const videosWithTechnicalFields = videos.map((video) => ({
            ...video,
            duration_seconds: 3599,
            height: 1080,
        }));
        const { rerender } = render(
            <VideoLibrarySection {...props} videos={videosWithTechnicalFields} />
        );

        expect(screen.getByText('59:59')).toBeInTheDocument();
        expect(screen.getByText('VIDEO_BADGE_RESOLUTION_FHD')).toBeInTheDocument();

        rerender(withQueryClient(<VideoLibrarySection {...props} videos={videos} />));

        expect(screen.queryByText('59:59')).not.toBeInTheDocument();
        expect(screen.queryByText(/VIDEO_BADGE_RESOLUTION/)).not.toBeInTheDocument();
    });
});
