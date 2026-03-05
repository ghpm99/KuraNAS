import { fireEvent, render, screen } from '@testing-library/react';
import VideoPlayerPage from './videoPlayer';

const mockUseVideoPlayer = jest.fn();
const mockNavigate = jest.fn();
const mockUseParams = jest.fn();
const mockUseLocation = jest.fn();
const mockPlayVideo = jest.fn();

jest.mock('@/components/hooks/useVideoPlayer/useVideoPlayer', () => ({
	__esModule: true,
	default: (...args: any[]) => mockUseVideoPlayer(...args),
}));

jest.mock('react-router-dom', () => ({
	useNavigate: () => mockNavigate,
	useParams: () => mockUseParams(),
	useLocation: () => mockUseLocation(),
}));

jest.mock('@/components/videos/videoControls/videoControls', () => ({
	__esModule: true,
	default: (props: any) => <div data-testid='video-controls'>{props.isPlaying ? 'playing' : 'paused'}</div>,
}));

jest.mock('@/components/videos/videoPlayer/videoPlayer', () => ({
	__esModule: true,
	default: (props: any) => (
		<div>
			<div data-testid='video-player'>{props.currentVideo?.name ?? 'no-video'}</div>
			<button type='button' onClick={props.onBack}>
				back
			</button>
		</div>
	),
}));
jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (k: string) => ({ VIDEO_INVALID_ID: 'Invalid video ID' }[k] ?? k) }),
}));

describe('pages/videoPlayer/videoPlayer', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockUseParams.mockReturnValue({ id: '30' });
		mockUseLocation.mockReturnValue({ state: { from: '/videos?playlist=x&video=30', playlistId: 7 } });
		mockUseVideoPlayer.mockReturnValue({
			videoRef: { current: null },
			playVideo: mockPlayVideo,
			seekTo: jest.fn(),
			setVolume: jest.fn(),
			setPlaybackRate: jest.fn(),
			toggleFullscreen: jest.fn(),
			togglePlayPause: jest.fn(),
			nextVideo: jest.fn(),
			previousVideo: jest.fn(),
			status: 'playing',
			currentTime: 12,
			duration: 33,
			volume: 0.8,
			playbackRate: 1.25,
			isFullscreen: false,
			setCurrentTime: jest.fn(),
			setDuration: jest.fn(),
			currentVideo: { id: 30, name: 'episode.mp4' },
		});
	});

	it('renders invalid state when route id is missing', () => {
		mockUseParams.mockReturnValue({ id: undefined });
		render(<VideoPlayerPage />);
		expect(screen.getByText('Invalid video ID')).toBeInTheDocument();
		expect(mockPlayVideo).not.toHaveBeenCalled();
	});

	it('plays on mount and navigates back to origin from state', () => {
		render(<VideoPlayerPage />);
		expect(screen.getByTestId('video-controls')).toHaveTextContent('playing');
		expect(screen.getByTestId('video-player')).toHaveTextContent('episode.mp4');
		expect(mockPlayVideo).toHaveBeenCalledTimes(1);

		fireEvent.click(screen.getByRole('button', { name: 'back' }));
		expect(mockNavigate).toHaveBeenCalledWith('/videos?playlist=x&video=30');
	});

	it('navigates history back when there is no from state', () => {
		window.history.pushState({}, '', '/before-video');
		mockUseLocation.mockReturnValue({ state: null });
		render(<VideoPlayerPage />);

		fireEvent.click(screen.getByRole('button', { name: 'back' }));
		expect(mockNavigate).toHaveBeenCalledWith(-1);
	});
});
