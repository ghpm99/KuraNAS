import { fireEvent, render, screen } from '@testing-library/react';
import type { VideoSection } from '@/app/routes';
import type { VideoPlaylistDto } from '@/service/videoPlayback';
import type { VideoContentContextData } from '@/features/videos/providers/videoContentProvider/videoContentProvider';

const mockUseVideoContentProvider = jest.fn();

jest.mock('@/features/videos/providers/videoContentProvider/videoContentProvider', () => ({
    __esModule: true,
    useVideoContentProvider: () => mockUseVideoContentProvider(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string) => key,
    }),
}));

jest.mock('./VideoSectionPlaylistGrid', () => ({
    __esModule: true,
    default: (props: { titleKey: string }) => (
        <div data-testid={`section-grid-${props.titleKey}`}>{props.titleKey}</div>
    ),
}));

jest.mock('./VideoContinueWatchingSection', () => ({
    __esModule: true,
    default: (props: { items: unknown[] }) => (
        <div data-testid="continue-section">{props.items.length}</div>
    ),
}));

jest.mock('./VideoLibrarySection', () => ({
    __esModule: true,
    default: (props: any) => (
        <div data-testid="library-section">
            {props.videos?.length ?? 0}-{props.playlists?.length ?? 0}
        </div>
    ),
}));

jest.mock('./VideoHomeScreen', () => ({
    __esModule: true,
    default: (props: any) => (
        <div data-testid="home-screen">
            {props.isLoadingPlaylists ? 'playlists-loading' : 'playlists-ready'}
            {props.isLoadingContinueWatching ? 'continue-loading' : 'continue-ready'}
            {props.isLoadingHomeCatalog ? 'catalog-loading' : 'catalog-ready'}
        </div>
    ),
}));

jest.mock('./VideoContextDetailView', () => ({
    __esModule: true,
    default: () => <div data-testid="context-detail">context</div>,
}));

jest.mock('./VideoSeriesDetailView', () => ({
    __esModule: true,
    default: (props: any) => (
        <div data-testid="series-detail">{props.playlist?.classification}</div>
    ),
}));

jest.mock('./VideoPlaylistDetailView', () => ({
    __esModule: true,
    default: () => <div data-testid="playlist-detail">playlist-detail</div>,
}));

jest.mock('./VideoFeedbackSnackbar', () => ({
    __esModule: true,
    default: (props: any) => (
        <div data-testid="feedback-snackbar">{props.open ? props.message : 'feedback-closed'}</div>
    ),
}));

import VideoContentScreen from './VideoContentScreen';

const createPlaylist = (overrides: Partial<VideoPlaylistDto> = {}): VideoPlaylistDto => ({
    id: overrides.id ?? 1,
    type: overrides.type ?? 'custom',
    source_path: overrides.source_path ?? '/videos',
    name: overrides.name ?? 'Playlist',
    is_hidden: overrides.is_hidden ?? false,
    is_auto: overrides.is_auto ?? false,
    group_mode: overrides.group_mode ?? 'single',
    classification: overrides.classification ?? 'movie',
    item_count: overrides.item_count ?? 0,
    cover_video_id: overrides.cover_video_id ?? null,
    created_at: overrides.created_at ?? '2023-01-01T00:00:00Z',
    updated_at: overrides.updated_at ?? '2023-01-01T00:00:00Z',
    last_played_at: overrides.last_played_at ?? null,
    items: overrides.items ?? [],
});

const createContext = (
    overrides: Partial<VideoContentContextData> = {}
): VideoContentContextData => ({
    currentSection: 'home',
    playlists: [],
    allVideos: [],
    filteredVideos: [],
    continueWatchingItems: [],
    seriesPlaylists: [],
    moviePlaylists: [],
    personalPlaylists: [],
    clipPlaylists: [],
    folderPlaylists: [],
    recentCatalogItems: [],
    playlistMembershipMap: {},
    selectedPlaylistSummary: null,
    selectedPlaylistDetail: null,
    isLoadingPlaylists: false,
    isLoadingVideos: false,
    isLoadingSelectedPlaylist: false,
    isLoadingHomeCatalog: false,
    isLoadingContinueWatching: false,
    playlistsFailure: null,
    videosFailure: null,
    selectedPlaylistFailure: null,
    homeCatalogFailure: null,
    continueWatchingFailure: null,
    isFetchingMoreVideos: false,
    hasMoreVideos: false,
    isAddingToPlaylist: false,
    isRenamingPlaylist: false,
    isRemovingFromPlaylist: false,
    isReorderingPlaylist: false,
    videoSearch: '',
    selectedPlaylistPerVideo: {},
    feedback: {
        open: false,
        message: '',
        severity: 'success',
    },
    setVideoSearch: jest.fn(),
    setSelectedPlaylistForVideo: jest.fn(),
    closeFeedback: jest.fn(),
    loadMoreVideos: jest.fn(),
    selectPlaylist: jest.fn(),
    clearSelectedPlaylist: jest.fn(),
    playVideo: jest.fn(),
    openPlaylistVideo: jest.fn(),
    addVideoFromLibrary: jest.fn(),
    renameSelectedPlaylist: jest.fn(),
    removeVideoFromSelectedPlaylist: jest.fn(),
    moveSelectedPlaylistItem: jest.fn(),
    setVideoWatched: jest.fn(),
    ...overrides,
});

const sectionTitleMap: Record<Exclude<VideoSection, 'home' | 'folders' | 'continue'>, string> = {
    series: 'VIDEO_SECTION_SERIES',
    movies: 'VIDEO_SECTION_MOVIES',
    personal: 'VIDEO_SECTION_PERSONAL',
    clips: 'VIDEO_SECTION_CLIPS',
};

describe('VideoContentScreen', () => {
    beforeEach(() => {
        mockUseVideoContentProvider.mockReset();
    });

    const renderScreen = (overrides: Partial<VideoContentContextData> = {}) => {
        mockUseVideoContentProvider.mockReturnValue(createContext(overrides));
        render(<VideoContentScreen />);
    };

    it('renders the page shell immediately while every query is pending', () => {
        renderScreen({
            isLoadingPlaylists: true,
            isLoadingVideos: true,
            isLoadingHomeCatalog: true,
            isLoadingContinueWatching: true,
        });
        expect(screen.getByTestId('feedback-snackbar')).toBeInTheDocument();
        expect(screen.getByTestId('home-screen')).toHaveTextContent(
            'playlists-loadingcontinue-loadingcatalog-loading'
        );
    });

    it.each(['series', 'movies', 'personal', 'clips'] as const)(
        'shows only a skeleton in the %s section while playlists load',
        (section) => {
            renderScreen({ currentSection: section, isLoadingPlaylists: true });
            expect(screen.getByTestId('video-section-skeleton')).toBeInTheDocument();
            expect(screen.queryByTestId(/^section-grid-/)).not.toBeInTheDocument();
        }
    );

    it('shows a skeleton in the continue section while its query loads', () => {
        renderScreen({ currentSection: 'continue', isLoadingContinueWatching: true });
        expect(screen.getByTestId('video-section-skeleton')).toBeInTheDocument();
        expect(screen.queryByTestId('continue-section')).not.toBeInTheDocument();
    });

    it('lets the folders sections resolve independently', () => {
        renderScreen({ currentSection: 'folders', isLoadingVideos: true });
        expect(screen.getByTestId('section-grid-VIDEO_SECTION_FOLDERS')).toBeInTheDocument();
        expect(screen.getByTestId('video-section-skeleton')).toBeInTheDocument();
        expect(screen.queryByTestId('library-section')).not.toBeInTheDocument();
    });

    it('shows the library while folder playlists are still loading', () => {
        renderScreen({ currentSection: 'folders', isLoadingPlaylists: true });
        expect(screen.getByTestId('video-section-skeleton')).toBeInTheDocument();
        expect(screen.getByTestId('library-section')).toBeInTheDocument();
    });

    it('replaces skeletons with content once the query resolves', () => {
        mockUseVideoContentProvider.mockReturnValue(
            createContext({ currentSection: 'series', isLoadingPlaylists: true })
        );
        const { rerender } = render(<VideoContentScreen />);
        expect(screen.getByTestId('video-section-skeleton')).toBeInTheDocument();
        mockUseVideoContentProvider.mockReturnValue(createContext({ currentSection: 'series' }));
        rerender(<VideoContentScreen />);
        expect(screen.queryByTestId('video-section-skeleton')).not.toBeInTheDocument();
        expect(screen.getByTestId('section-grid-VIDEO_SECTION_SERIES')).toBeInTheDocument();
    });

    it('shows the playlist loader while a playlist detail is loading', () => {
        const playlist = createPlaylist({ id: 2 });
        renderScreen({
            selectedPlaylistSummary: playlist,
            selectedPlaylistDetail: null,
            isLoadingSelectedPlaylist: true,
            currentSection: 'series',
        });
        expect(screen.getByText('VIDEO_LOADING_PLAYLIST')).toBeInTheDocument();
    });

    it('renders the series detail view when the playlist is marked as series', () => {
        const playlist = createPlaylist({ classification: 'series' });
        renderScreen({
            selectedPlaylistSummary: playlist,
            selectedPlaylistDetail: playlist,
            currentSection: 'series',
        });
        expect(screen.getByTestId('series-detail')).toBeInTheDocument();
    });

    it('renders the context detail view for non-folder playlists', () => {
        const playlist = createPlaylist({ classification: 'movie' });
        renderScreen({
            selectedPlaylistSummary: playlist,
            selectedPlaylistDetail: playlist,
            currentSection: 'movies',
        });
        expect(screen.getByTestId('context-detail')).toBeInTheDocument();
    });

    it('renders the playlist detail view when the section is folders', () => {
        const playlist = createPlaylist({ classification: 'movie' });
        renderScreen({
            selectedPlaylistSummary: playlist,
            selectedPlaylistDetail: playlist,
            currentSection: 'folders',
        });
        expect(screen.getByTestId('playlist-detail')).toBeInTheDocument();
    });

    it.each(Object.entries(sectionTitleMap))('renders the %s section grid', (section, titleKey) => {
        renderScreen({ currentSection: section as VideoSection });
        expect(screen.getByTestId(`section-grid-${titleKey}`)).toBeInTheDocument();
    });

    it('renders the continue watching section with the in-progress videos', () => {
        renderScreen({
            currentSection: 'continue',
            continueWatchingItems: [
                {
                    video: {
                        id: 1,
                        name: 'a',
                        path: '/a',
                        parent_path: '/',
                        format: '.mp4',
                        size: 1,
                    },
                    position_seconds: 5,
                    duration_seconds: 10,
                    updated_at: '2026-01-01T00:00:00Z',
                },
            ],
        });
        expect(screen.getByTestId('continue-section')).toHaveTextContent('1');
    });

    it('renders the folders section with the library section', () => {
        renderScreen({ currentSection: 'folders' });
        expect(screen.getByTestId('section-grid-VIDEO_SECTION_FOLDERS')).toBeInTheDocument();
        expect(screen.getByTestId('library-section')).toBeInTheDocument();
    });

    it('renders the home screen for the home section', () => {
        renderScreen({ currentSection: 'home' });
        expect(screen.getByTestId('home-screen')).toBeInTheDocument();
    });

    it.each(['series', 'movies', 'personal', 'clips'] as const)(
        'shows an error with retry in the %s section when playlists fail',
        (section) => {
            const retry = jest.fn();
            renderScreen({
                currentSection: section,
                playlistsFailure: { message: 'backend down', retry },
            });
            expect(screen.getByText('backend down')).toBeInTheDocument();
            expect(screen.queryByTestId(/^section-grid-/)).not.toBeInTheDocument();
            fireEvent.click(screen.getByRole('button'));
            expect(retry).toHaveBeenCalledTimes(1);
        }
    );

    it('shows an error with retry in the continue section when its query fails', () => {
        const retry = jest.fn();
        renderScreen({
            currentSection: 'continue',
            continueWatchingFailure: { message: 'continue down', retry },
        });
        expect(screen.getByText('continue down')).toBeInTheDocument();
        expect(screen.queryByTestId('continue-section')).not.toBeInTheDocument();
        fireEvent.click(screen.getByRole('button'));
        expect(retry).toHaveBeenCalledTimes(1);
    });

    it('keeps the folder playlists when only the library query fails', () => {
        const retry = jest.fn();
        renderScreen({
            currentSection: 'folders',
            videosFailure: { message: 'library down', retry },
        });
        expect(screen.getByTestId('section-grid-VIDEO_SECTION_FOLDERS')).toBeInTheDocument();
        expect(screen.getByText('library down')).toBeInTheDocument();
        expect(screen.queryByTestId('library-section')).not.toBeInTheDocument();
        fireEvent.click(screen.getByRole('button'));
        expect(retry).toHaveBeenCalledTimes(1);
    });

    it('keeps the library when only the folder playlists query fails', () => {
        renderScreen({
            currentSection: 'folders',
            playlistsFailure: { message: 'playlists down', retry: jest.fn() },
        });
        expect(screen.getByText('playlists down')).toBeInTheDocument();
        expect(screen.getByTestId('library-section')).toBeInTheDocument();
    });

    it('shows an error with retry when the selected playlist detail fails', () => {
        const retry = jest.fn();
        renderScreen({
            selectedPlaylistSummary: createPlaylist({ id: 2 }),
            selectedPlaylistFailure: { message: 'detail down', retry },
            currentSection: 'series',
        });
        expect(screen.getByText('detail down')).toBeInTheDocument();
        expect(screen.queryByText('VIDEO_LOADING_PLAYLIST')).not.toBeInTheDocument();
        fireEvent.click(screen.getByRole('button'));
        expect(retry).toHaveBeenCalledTimes(1);
    });

    it('passes the per-query failures to the home screen', () => {
        renderScreen({ currentSection: 'home' });
        expect(screen.getByTestId('home-screen')).toBeInTheDocument();
    });

    it('shows the feedback snackbar when feedback is open', () => {
        renderScreen({
            feedback: {
                open: true,
                message: 'Video ready',
                severity: 'success',
            },
        });
        expect(screen.getByTestId('feedback-snackbar')).toHaveTextContent('Video ready');
    });
});
