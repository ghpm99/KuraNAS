import { render, screen } from '@testing-library/react';
import type { VideoSection } from '@/app/routes';
import type { VideoPlaylistDto } from '@/service/videoPlayback';
import type { VideoContentContextData } from '@/components/providers/videoContentProvider/videoContentProvider';

const mockUseVideoContentProvider = jest.fn();

jest.mock('@/components/providers/videoContentProvider/videoContentProvider', () => ({
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
    default: () => <div data-testid="home-screen">home</div>,
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
    continuePlaylists: [],
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
    ...overrides,
});

const sectionTitleMap: Record<Exclude<VideoSection, 'home' | 'folders'>, string> = {
    continue: 'VIDEO_SECTION_CONTINUE',
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

    it('shows the video loader when queries are running', () => {
        renderScreen({ isLoadingPlaylists: true });
        expect(screen.getByText('VIDEO_LOADING_VIDEOS')).toBeInTheDocument();
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

    it('renders the folders section with the library section', () => {
        renderScreen({ currentSection: 'folders' });
        expect(screen.getByTestId('section-grid-VIDEO_SECTION_FOLDERS')).toBeInTheDocument();
        expect(screen.getByTestId('library-section')).toBeInTheDocument();
    });

    it('renders the home screen for the home section', () => {
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
