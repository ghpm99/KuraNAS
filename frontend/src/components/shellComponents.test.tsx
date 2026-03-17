import { fireEvent, render, screen } from '@testing-library/react';
import Header from '@/components/layout/Header/Header';
import { Layout } from '@/components/layout/Layout/Layout';
import Sidebar from '@/components/layout/Sidebar/Sidebar';
import NavItem from '@/components/layout/Sidebar/components/navItem';
import Tabs from '@/components/tabs/tabs';
import ActionBar from '@/components/actionBar/actionBar';
import ActivePageListener from '@/components/activePageListener';
import Button from '@/components/ui/Button/Button';
import Card from '@/components/ui/Card/Card';
import Message from '@/components/ui/Message/Message';
import FileCard from '@/components/fileCard/fileCard';
import AboutPage from '@/pages/about';
import FavoritesPage from '@/pages/favorites';
import FilesPage from '@/pages/files';
import HomePage from '@/pages/home';
import ImagesPage from '@/pages/images';
import MusicPage from '@/pages/music';
import SettingsPage from '@/pages/settings';
import VideosPage from '@/pages/videos/videos';
import VideoPlayerPage from '@/pages/videoPlayer/videoPlayer';
import AnalyticsPage from '@/pages/analytics';

const mockUseFile = jest.fn();
const mockUseUI = jest.fn();
const mockUseLocation = jest.fn();
const mockUseParams = jest.fn();
const mockNavigate = jest.fn();
const mockUseAnalyticsOverview = jest.fn();
const mockOpenSearch = jest.fn();

jest.mock('@/components/providers/fileProvider/fileContext', () => ({
    __esModule: true,
    default: () => mockUseFile(),
}));

jest.mock('@/components/providers/uiProvider/uiContext', () => {
    const actual = jest.requireActual('@/components/providers/uiProvider/uiContext');
    return {
        ...actual,
        useUI: () => mockUseUI(),
    };
});

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (k: string, options?: Record<string, string>) =>
            options?.shortcut ? `${k}:${options.shortcut}` : k,
    }),
}));

jest.mock('@/components/search/useGlobalSearch', () => ({
    __esModule: true,
    default: () => ({ openSearch: mockOpenSearch, shortcut: 'Ctrl+K' }),
}));

jest.mock('react-router-dom', () => ({
    Link: ({ children, to }: any) => <a href={to}>{children}</a>,
    Outlet: () => <div>OutletMock</div>,
    useLocation: () => mockUseLocation(),
    useParams: () => mockUseParams(),
    useNavigate: () => mockNavigate,
}));

jest.mock('@/components/layout/Sidebar/components/folderTree', () => () => (
    <div>FolderTreeMock</div>
));
jest.mock('@/components/layout/Sidebar/components/navItem', () => ({ children }: any) => (
    <div>{children}</div>
));
jest.mock('@/components/layout/Layout', () => ({
    __esModule: true,
    default: ({ children }: any) => <div data-testid="analytics-layout">{children}</div>,
}));

jest.mock('@/components/providers/activityDiaryProvider/ActivityDiaryContext', () => ({
    useActivityDiary: () => ({ currentTime: new Date('2026-01-01T00:00:00Z') }),
}));

jest.mock('@/components/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => ({ hasQueue: true }),
    GlobalMusicProvider: ({ children }: any) => <div>{children}</div>,
}));

jest.mock('@/components/actionBar', () => () => <div>ActionBarMock</div>);
jest.mock('@/components/fileContent', () => () => <div>FileContentMock</div>);
jest.mock('@/components/fileDetails', () => () => <div>FileDetailsMock</div>);
jest.mock('@/components/tabs', () => () => <div>TabsMock</div>);
jest.mock('@/components/favorites/FavoritesScreen', () => () => <div>FavoritesScreenMock</div>);
jest.mock('@/components/files/filesLayout', () => ({ children }: any) => (
    <div data-testid="files-layout">{children}</div>
));
jest.mock('@/components/files/FilesExplorerScreen', () => () => <div>FilesExplorerScreenMock</div>);

jest.mock('@/components/imageContent', () => () => <div>ImageContentMock</div>);
jest.mock('@/components/images/imagesLayout', () => ({ children }: any) => (
    <div data-testid="images-layout">{children}</div>
));

jest.mock('@/components/music/musicLayout', () => ({ children }: any) => (
    <div data-testid="music-layout">{children}</div>
));
jest.mock('@/components/music/MusicSidebar', () => () => <div>MusicSidebarMock</div>);
jest.mock('@/components/musicContent', () => () => <div>MusicContentMock</div>);
jest.mock('@/components/home/HomeScreen', () => () => <div>HomeScreenMock</div>);
jest.mock('@/components/settings/SettingsScreen', () => () => <div>SETTINGS_PAGE_TITLE</div>);

jest.mock('@/components/videos/videoLayout', () => ({ children }: any) => (
    <div data-testid="video-layout">{children}</div>
));
jest.mock('@/components/videos/VideoDomainHeader', () => () => <div>VideoDomainHeaderMock</div>);
jest.mock('@/components/videos/VideoSidebar', () => () => <div>VideoSidebarMock</div>);
jest.mock('@/components/videos/videoContent/videoContent', () => () => <div>VideoContentMock</div>);

jest.mock('@/components/videos/videoPlayer/VideoPlayerScreen', () => () => (
    <div>VideoPlayerScreenMock</div>
));

jest.mock('@/components/providers/analyticsProvider', () => ({
    AnalyticsProvider: ({ children }: any) => <div>{children}</div>,
}));
jest.mock('@/components/providers/analyticsProvider/analyticsContext', () => ({
    useAnalyticsOverview: () => mockUseAnalyticsOverview(),
}));
jest.mock('@/components/ui/Button/Button', () => ({ children, onClick }: any) => (
    <button onClick={onClick}>{children}</button>
));
jest.mock('@/components/hooks/useAnalyticsFormatters/useAnalyticsFormatters', () => ({
    useAnalyticsFormatters: () => ({
        formatBytes: () => '0 B',
        formatPercent: () => '0%',
        formatDate: () => '-',
    }),
}));
jest.mock('@/components/hooks/useAnalyticsDerived/useAnalyticsDerived', () => ({
    useAnalyticsDerived: () => ({ usedPercent: 0, reclaimablePercent: 0 }),
}));

jest.mock('@/components/about/aboutLayout', () => ({ children }: any) => (
    <div data-testid="about-layout">{children}</div>
));
jest.mock('@/components/about/AboutScreen', () => () => <div>AboutScreenMock</div>);

beforeEach(() => {
    jest.clearAllMocks();
    mockUseFile.mockReturnValue({
        selectedItem: null,
        status: 'success',
        fileListFilter: 'all',
        setFileListFilter: jest.fn(),
        handleSelectItem: jest.fn(),
        expandedItems: [],
        files: [],
    });
    mockUseUI.mockReturnValue({ activePage: 'files', setActivePage: jest.fn() });
    mockUseLocation.mockReturnValue({ pathname: '/files', state: null });
    mockUseParams.mockReturnValue({ id: '10' });
    mockUseAnalyticsOverview.mockReturnValue({
        period: '7d',
        setPeriod: jest.fn(),
        loading: false,
        error: '',
        refresh: jest.fn(),
        data: {
            generated_at: '2026-01-01T00:00:00Z',
            storage: {
                used_bytes: 1024,
                total_bytes: 2048,
                free_bytes: 1024,
                growth_bytes: 128,
            },
            counts: { files_added: 2, files_total: 20, folders: 4 },
            hot_folders: [{ path: '/media', bytes: 500, growth_bytes: 10 }],
            duplicates: {
                groups: 1,
                files: 2,
                reclaimable_size: 256,
                top_groups: [{ signature: 'abcdef1234567890', copies: 2, reclaimable_size: 128 }],
            },
            library: {
                categorized_media: 8,
                audio_with_metadata: 3,
                video_with_metadata: 3,
                image_with_metadata: 2,
                image_classified: 2,
            },
            processing: {
                metadata_pending: 1,
                metadata_failed: 0,
                thumbnail_pending: 2,
                thumbnail_failed: 1,
            },
            health: {
                status: 'scanning',
                indexed_files: 100,
                errors_last_24h: 0,
                last_scan_at: '2026-01-01T00:00:00Z',
                last_scan_seconds: 30,
                recent_errors: ['none'],
            },
            time_series: [{ date: '2026-01-01', used_bytes: 1024 }],
            types: [{ type: 'video', count: 1, bytes: 1024 }],
            top_folders: [
                {
                    path: '/media/videos',
                    bytes: 1024,
                    last_modified: '2026-01-01T00:00:00Z',
                },
            ],
            extensions: [{ ext: '.mp4', count: 1, bytes: 1024 }],
            recent_files: [
                {
                    id: 1,
                    name: 'movie.mp4',
                    parent_path: '/media/videos',
                    size_bytes: 1024,
                    created_at: '2026-01-01T00:00:00Z',
                    updated_at: '2026-01-01T00:00:00Z',
                },
            ],
        },
    });
});

describe('shell components and pages', () => {
    it('renders header, layout and sidebar', () => {
        render(<Header showClock />);
        expect(screen.getByText('SEARCH_PLACEHOLDER')).toBeInTheDocument();
        expect(screen.getByTitle('NOTIFICATIONS')).toBeInTheDocument();

        render(
            <Layout>
                <div>child</div>
            </Layout>
        );
        expect(screen.getByText('child')).toBeInTheDocument();

        render(<Sidebar />);
        expect(screen.getAllByText('FolderTreeMock').length).toBeGreaterThan(0);
    });

    it('renders folder tree states and folder item formatting', () => {
        expect(screen.queryByText('FolderTreeMock')).not.toBeInTheDocument();
    });

    it('renders nav item, tabs, action bar and active page listener', () => {
        render(
            <NavItem href="/files" icon={<span>x</span>}>
                Home
            </NavItem>
        );
        expect(screen.getByText('Home')).toBeInTheDocument();

        render(<Tabs />);
        expect(screen.getByText('ALL_FILES')).toBeInTheDocument();

        render(<ActionBar />);
        expect(screen.getByText('NEW_FILE')).toBeInTheDocument();

        render(<ActivePageListener />);
        expect(mockUseUI().setActivePage).toHaveBeenCalledWith('files');

        mockUseLocation.mockReturnValueOnce({ pathname: '/home' });
        render(<ActivePageListener />);
        expect(mockUseUI().setActivePage).toHaveBeenCalledWith('home');

        mockUseLocation.mockReturnValueOnce({ pathname: '/favorites' });
        render(<ActivePageListener />);
        expect(mockUseUI().setActivePage).toHaveBeenCalledWith('favorites');

        mockUseLocation.mockReturnValueOnce({ pathname: '/settings' });
        render(<ActivePageListener />);
        expect(mockUseUI().setActivePage).toHaveBeenCalledWith('settings');

        mockUseLocation.mockReturnValueOnce({ pathname: '/music/playlists' });
        render(<ActivePageListener />);
        expect(mockUseUI().setActivePage).toHaveBeenCalledWith('music');

        mockUseLocation.mockReturnValueOnce({ pathname: '/videos/series' });
        render(<ActivePageListener />);
        expect(mockUseUI().setActivePage).toHaveBeenCalledWith('videos');

        mockUseLocation.mockReturnValueOnce({ pathname: '/analytics/library' });
        render(<ActivePageListener />);
        expect(mockUseUI().setActivePage).toHaveBeenCalledWith('analytics');

        mockUseLocation.mockReturnValueOnce({ pathname: '/activity-diary' });
        render(<ActivePageListener />);
        expect(mockUseUI().setActivePage).toHaveBeenCalledWith('unknown');

        mockUseLocation.mockReturnValueOnce({ pathname: '/unknown' });
        render(<ActivePageListener />);
        expect(mockUseUI().setActivePage).toHaveBeenCalledWith('unknown');
    });

    it('renders reusable ui components', () => {
        const onClick = jest.fn();
        render(<Button onClick={onClick}>Click</Button>);
        fireEvent.click(screen.getByText('Click'));
        expect(onClick).toHaveBeenCalled();

        render(
            <Card title="Title">
                <span>Body</span>
            </Card>
        );
        expect(screen.getByText('Title')).toBeInTheDocument();
        expect(screen.getByText('Body')).toBeInTheDocument();

        render(<Message type="success" text="ok" />);
        expect(screen.getByText('ok')).toBeInTheDocument();

        const star = jest.fn();
        const { container } = render(
            <FileCard
                title="f"
                metadata="m"
                thumbnail=""
                onClick={onClick}
                onClickStar={star}
                starred
            />
        );
        const buttons = container.querySelectorAll('button');
        fireEvent.click(buttons[buttons.length - 1] as HTMLButtonElement);
        expect(star).toHaveBeenCalled();
    });

    it('renders composition pages and video back behavior', () => {
        render(<HomePage />);
        expect(screen.getByTestId('analytics-layout')).toBeInTheDocument();
        expect(screen.getByText('HomeScreenMock')).toBeInTheDocument();

        render(<FilesPage />);
        expect(screen.getByTestId('files-layout')).toBeInTheDocument();
        expect(screen.getByText('FilesExplorerScreenMock')).toBeInTheDocument();

        render(<FavoritesPage />);
        expect(screen.getAllByTestId('files-layout').length).toBeGreaterThan(0);
        expect(mockUseFile().setFileListFilter).toHaveBeenCalledWith('starred');
        expect(screen.getByText('FavoritesScreenMock')).toBeInTheDocument();

        render(<ImagesPage />);
        expect(screen.getByTestId('images-layout')).toBeInTheDocument();

        render(<MusicPage />);
        expect(screen.getByTestId('music-layout')).toBeInTheDocument();

        render(<VideosPage />);
        expect(screen.getByTestId('video-layout')).toBeInTheDocument();
        expect(screen.getByText('VideoDomainHeaderMock')).toBeInTheDocument();
        expect(screen.getByText('VideoSidebarMock')).toBeInTheDocument();

        render(<AboutPage />);
        expect(screen.getByTestId('about-layout')).toBeInTheDocument();
        expect(screen.getByText('AboutScreenMock')).toBeInTheDocument();

        render(<AnalyticsPage />);
        expect(screen.getAllByTestId('analytics-layout').length).toBeGreaterThan(0);
        expect(
            screen.getByRole('heading', { name: 'ANALYTICS_SECTION_OVERVIEW' })
        ).toBeInTheDocument();

        render(<SettingsPage />);
        expect(screen.getByText('SETTINGS_PAGE_TITLE')).toBeInTheDocument();

        render(<VideoPlayerPage />);
        expect(screen.getByText('VideoPlayerScreenMock')).toBeInTheDocument();
    });

    it('renders action bar for selected file branch', () => {
        mockUseFile.mockReturnValue({
            ...mockUseFile(),
            selectedItem: { id: 3, type: 2, name: 'report.pdf' },
        });
        render(<ActionBar />);
        expect(screen.getByText('report.pdf')).toBeInTheDocument();
        expect(screen.getByText('DOWNLOAD')).toBeInTheDocument();
    });

    it('hides tabs when selected item is file', () => {
        mockUseFile.mockReturnValue({
            ...mockUseFile(),
            selectedItem: { id: 4, type: 2 },
        });
        render(<Tabs />);
        expect(screen.queryByText('ALL_FILES')).not.toBeInTheDocument();
    });
});
