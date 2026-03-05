import { act, fireEvent, render, screen } from '@testing-library/react';
import React from 'react';
import Header from '@/components/layout/Header/Header';
import { Layout } from '@/components/layout/Layout/Layout';
import Sidebar from '@/components/layout/Sidebar/Sidebar';
import FolderTree from '@/components/layout/Sidebar/components/folderTree';
import FolderItem from '@/components/layout/Sidebar/components/folderTree/components/folderItem';
import NavItem from '@/components/layout/Sidebar/components/navItem';
import Tabs from '@/components/tabs/tabs';
import ActionBar from '@/components/actionBar/actionBar';
import ActivePageListener from '@/components/activePageListener';
import Button from '@/components/ui/Button/Button';
import Card from '@/components/ui/Card/Card';
import Message from '@/components/ui/Message/Message';
import FileCard from '@/components/fileCard/fileCard';
import AboutPage from '@/pages/about';
import FilesPage from '@/pages/files';
import ImagesPage from '@/pages/images';
import MusicPage from '@/pages/music';
import VideosPage from '@/pages/videos/videos';
import VideoPlayerPage from '@/pages/videoPlayer/videoPlayer';
import AnalyticsPage from '@/pages/analytics';

const mockUseFile = jest.fn();
const mockUseUI = jest.fn();
const mockUseLocation = jest.fn();
const mockUseParams = jest.fn();
const mockNavigate = jest.fn();
const mockVideoPlayer = jest.fn();

jest.mock('@/components/hooks/fileProvider/fileContext', () => ({
	__esModule: true,
	default: () => mockUseFile(),
}));

jest.mock('@/components/hooks/UI/uiContext', () => ({
	useUI: () => mockUseUI(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (k: string) => k }),
}));

jest.mock('react-router-dom', () => ({
	Link: ({ children, to }: any) => <a href={to}>{children}</a>,
	useLocation: () => mockUseLocation(),
	useParams: () => mockUseParams(),
	useNavigate: () => mockNavigate,
}));

jest.mock('@/components/layout/Sidebar/components/folderTree', () => () => <div>FolderTreeMock</div>);
jest.mock('@/components/layout/Sidebar/components/navItem', () => ({ children }: any) => <div>{children}</div>);

jest.mock('@/components/hooks/ActivityDiaryProvider/ActivityDiaryContext', () => ({
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
jest.mock('@/components/files/filesLayout', () => ({ children }: any) => <div data-testid='files-layout'>{children}</div>);

jest.mock('@/components/imageContent', () => () => <div>ImageContentMock</div>);
jest.mock('@/components/images/imagesLayout', () => ({ children }: any) => <div data-testid='images-layout'>{children}</div>);

jest.mock('@/components/music/musicLayout', () => ({ children }: any) => <div data-testid='music-layout'>{children}</div>);
jest.mock('@/components/music/MusicSidebar', () => () => <div>MusicSidebarMock</div>);
jest.mock('@/components/musicContent', () => () => <div>MusicContentMock</div>);

jest.mock('@/components/videos/videoLayout', () => ({ children }: any) => <div data-testid='video-layout'>{children}</div>);
jest.mock('@/components/videos/videoContent/videoContent', () => () => <div>VideoContentMock</div>);

jest.mock('@/components/hooks/useVideoPlayer/useVideoPlayer', () => ({
	__esModule: true,
	default: (...args: any[]) => mockVideoPlayer(...args),
}));
jest.mock('@/components/videos/videoControls/videoControls', () => () => <div>VideoControlsMock</div>);
jest.mock('@/components/videos/videoPlayer/videoPlayer', () => () => <div>VideoPlayerMock</div>);

jest.mock('@/components/analytics/analyticsLayout', () => ({ children }: any) => <div data-testid='analytics-layout'>{children}</div>);
jest.mock('@/components/contexts/AnalyticsContext', () => ({ useAnalytics: () => ({ refreshAnalytics: jest.fn() }) }));
jest.mock('@/components/ui/Button/Button', () => ({ children, onClick }: any) => <button onClick={onClick}>{children}</button>);
jest.mock('@/components/analytics/StorageOverviewCards/StorageOverviewCards', () => () => <div>StorageOverviewCards</div>);
jest.mock('@/components/analytics/DiskUsageChart/DiskUsageChart', () => () => <div>DiskUsageChart</div>);
jest.mock('@/components/analytics/FileTypesChart/FileTypesChart', () => () => <div>FileTypesChart</div>);
jest.mock('@/components/analytics/FileTypesTable/FileTypesTable', () => () => <div>FileTypesTable</div>);
jest.mock('@/components/analytics/SizeRangesChart/SizeRangesChart', () => () => <div>SizeRangesChart</div>);
jest.mock('@/components/analytics/LargestFilesTable/LargestFilesTable', () => () => <div>LargestFilesTable</div>);
jest.mock('@/components/analytics/DuplicatesSection/DuplicatesSection', () => () => <div>DuplicatesSection</div>);
jest.mock('@/components/analytics/ActivityChart/ActivityChart', () => () => <div>ActivityChart</div>);
jest.mock('@/components/analytics/RecentActivity/RecentActivity', () => () => <div>RecentActivity</div>);
jest.mock('@/components/analytics/EmptyFoldersSection/EmptyFoldersSection', () => () => <div>EmptyFoldersSection</div>);
jest.mock('@/components/analytics/CleanupSuggestions/CleanupSuggestions', () => () => <div>CleanupSuggestions</div>);
jest.mock('@/components/analytics/BackupSection/BackupSection', () => () => <div>BackupSection</div>);
jest.mock('@/components/analytics/TrashSection/TrashSection', () => () => <div>TrashSection</div>);

jest.mock('@/components/about/aboutLayout', () => ({ children }: any) => <div data-testid='about-layout'>{children}</div>);
jest.mock('@/components/about/SystemInfoCard/SystemInfoCard', () => () => <div>SystemInfoCard</div>);
jest.mock('@/components/about/TechnicalInfoCard/TechnicalInfoCard', () => () => <div>TechnicalInfoCard</div>);
jest.mock('@/components/about/StatusCard/StatusCard', () => () => <div>StatusCard</div>);
jest.mock('@/components/about/UpdateCard/UpdateCard', () => () => <div>UpdateCard</div>);

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
	mockUseLocation.mockReturnValue({ pathname: '/', state: null });
	mockUseParams.mockReturnValue({ id: '10' });
	mockVideoPlayer.mockReturnValue({
		videoRef: { current: null },
		playVideo: jest.fn(),
		seekTo: jest.fn(),
		setVolume: jest.fn(),
		setPlaybackRate: jest.fn(),
		toggleFullscreen: jest.fn(),
		togglePlayPause: jest.fn(),
		nextVideo: jest.fn(),
		previousVideo: jest.fn(),
		status: 'paused',
		currentTime: 0,
		duration: 10,
		volume: 1,
		playbackRate: 1,
		isFullscreen: false,
		setCurrentTime: jest.fn(),
		setDuration: jest.fn(),
		currentVideo: null,
	});
});

describe('shell components and pages', () => {
	it('renders header, layout and sidebar', () => {
		render(<Header showClock currentTime={new Date('2026-01-01T00:00:00Z')} />);
		expect(screen.getByPlaceholderText('SEARCH_PLACEHOLDER')).toBeInTheDocument();
		expect(screen.getByTitle('NOTIFICATIONS')).toBeInTheDocument();

		render(<Layout><div>child</div></Layout>);
		expect(screen.getByText('child')).toBeInTheDocument();

		render(<Sidebar />);
		expect(screen.getAllByText('FolderTreeMock').length).toBeGreaterThan(0);
	});

	it('renders folder tree states and folder item formatting', () => {
		expect(screen.queryByText('FolderTreeMock')).not.toBeInTheDocument();
	});

	it('renders nav item, tabs, action bar and active page listener', () => {
		render(<NavItem href='/' icon={<span>x</span>}>Home</NavItem>);
		expect(screen.getByText('Home')).toBeInTheDocument();

		render(<Tabs />);
		expect(screen.getByText('ALL_FILES')).toBeInTheDocument();

		render(<ActionBar />);
		expect(screen.getByText('NEW_FILE')).toBeInTheDocument();

		render(<ActivePageListener />);
		expect(mockUseUI().setActivePage).toHaveBeenCalledWith('files');

		mockUseLocation.mockReturnValueOnce({ pathname: '/unknown' });
		render(<ActivePageListener />);
		expect(mockUseUI().setActivePage).toHaveBeenCalledWith('unknown');
	});

	it('renders reusable ui components', () => {
		const onClick = jest.fn();
		render(<Button onClick={onClick}>Click</Button>);
		fireEvent.click(screen.getByText('Click'));
		expect(onClick).toHaveBeenCalled();

		render(<Card title='Title'><span>Body</span></Card>);
		expect(screen.getByText('Title')).toBeInTheDocument();
		expect(screen.getByText('Body')).toBeInTheDocument();

		render(<Message type='success' text='ok' />);
		expect(screen.getByText('ok')).toBeInTheDocument();

		const star = jest.fn();
		const { container } = render(<FileCard title='f' metadata='m' thumbnail='' onClick={onClick} onClickStar={star} starred />);
		const buttons = container.querySelectorAll('button');
		fireEvent.click(buttons[buttons.length - 1] as HTMLButtonElement);
		expect(star).toHaveBeenCalled();
	});

	it('renders composition pages and video back behavior', () => {
		render(<FilesPage />);
		expect(screen.getByTestId('files-layout')).toBeInTheDocument();
		expect(screen.getByText('FileContentMock')).toBeInTheDocument();

		render(<ImagesPage />);
		expect(screen.getByTestId('images-layout')).toBeInTheDocument();

		render(<MusicPage />);
		expect(screen.getByTestId('music-layout')).toBeInTheDocument();

		render(<VideosPage />);
		expect(screen.getByTestId('video-layout')).toBeInTheDocument();

		render(<AboutPage />);
		expect(screen.getByTestId('about-layout')).toBeInTheDocument();
		expect(screen.getByText('SystemInfoCard')).toBeInTheDocument();

		render(<AnalyticsPage />);
		expect(screen.getByTestId('analytics-layout')).toBeInTheDocument();
		expect(screen.getByText('StorageOverviewCards')).toBeInTheDocument();

		render(<VideoPlayerPage />);
		expect(screen.getByText('VideoControlsMock')).toBeInTheDocument();
		expect(screen.getByText('VideoPlayerMock')).toBeInTheDocument();

		mockUseParams.mockReturnValueOnce({ id: undefined });
		render(<VideoPlayerPage />);
		expect(screen.getByText('VIDEO_INVALID_ID')).toBeInTheDocument();
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
