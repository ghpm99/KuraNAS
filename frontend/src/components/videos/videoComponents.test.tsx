import { act, fireEvent, render, screen } from '@testing-library/react';
import { createRef } from 'react';
import VideoControls from './videoControls/videoControls';
import VideoSettings from './videoSettings/videoSettings';
import VideoProgressBar from './videoProgressBar/videoProgressBar';
import VideoThumbnail from './videoThumbnail/videoThumbnail';
import VideoPlayer from './videoPlayer/videoPlayer';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string, params?: Record<string, string | number>) => {
			const map: Record<string, string> = {
				VIDEO_PLAYBACK_SPEED: 'Playback Speed',
				VIDEO_QUALITY: 'Quality',
				VIDEO_MORE_OPTIONS: 'More Options',
				VIDEO_SUBTITLES: 'Subtitles',
				VIDEO_NOT_AVAILABLE_YET: 'Not available yet',
				VIDEO_ASPECT_RATIO: 'Aspect Ratio',
				VIDEO_AUTO_DETECTED: 'Auto detected',
				VIDEO_AUDIO_TRACK: 'Audio Track',
				VIDEO_DEFAULT: 'Default',
				VIDEO_UNKNOWN: 'Unknown',
				VIDEO_NO_VIDEO_PLAYING: 'No video playing',
				VIDEO_BACK: 'Voltar',
				VIDEO_AUTO: 'Auto',
				VIDEO_NORMAL: 'Normal',
				VIDEO_FULLSCREEN: 'Fullscreen',
				VIDEO_EXIT_FULLSCREEN: 'Exit Fullscreen',
			};
			if (key === 'VIDEO_PLAY_ARIA') return `Play ${params?.title ?? ''}`.trim();
			if (key === 'VIDEO_CODEC_TOOLTIP') return `Codec: ${params?.codec ?? ''}`.trim();
			return map[key] ?? key;
		},
	}),
}));

describe('video components', () => {
	it('renders and interacts with video controls', () => {
		jest.useFakeTimers();
		const clearTimers = () => {
			jest.runOnlyPendingTimers();
			jest.useRealTimers();
		};
		const props = {
			isPlaying: true,
			currentTime: 20,
			duration: 100,
			volume: 0.5,
			playbackRate: 1,
			isFullscreen: false,
			seekTo: jest.fn(),
			setVolume: jest.fn(),
			setPlaybackRate: jest.fn(),
			toggleFullscreen: jest.fn(),
			togglePlayPause: jest.fn(),
			nextVideo: jest.fn(),
			previousVideo: jest.fn(),
		};
		const { container, unmount } = render(<VideoControls {...props} />);
		const playPauseButton = container.querySelector('.play-pause-button') as HTMLButtonElement;
		const skipButtons = container.querySelectorAll('.skip-button');
		const rightControlsButtons = container.querySelectorAll('.right-controls button');
		fireEvent.click(skipButtons[0]!);
		expect(props.seekTo).toHaveBeenCalledWith(10);
		fireEvent.click(skipButtons[1]!);
		expect(props.seekTo).toHaveBeenCalledWith(30);
		fireEvent.click(playPauseButton);
		expect(props.togglePlayPause).toHaveBeenCalled();

			fireEvent.click(rightControlsButtons[rightControlsButtons.length - 1]!);
		expect(props.toggleFullscreen).toHaveBeenCalled();
		fireEvent.click(container.querySelectorAll('.left-controls button')[0]!);
		fireEvent.click(container.querySelectorAll('.left-controls button')[1]!);
		expect(props.previousVideo).toHaveBeenCalled();
		expect(props.nextVideo).toHaveBeenCalled();

		const sliders = screen.getAllByRole('slider');
		fireEvent.change(sliders[0]!, { target: { value: '42' } });
		fireEvent.change(sliders[1]!, { target: { value: '0.7' } });
		expect(props.seekTo).toHaveBeenCalledWith(42);
		expect(props.setVolume).toHaveBeenCalledWith(0.7);
		fireEvent.mouseMove(document);
		fireEvent.keyPress(document, { key: 'k' });

		fireEvent.click(rightControlsButtons[1]!);
		fireEvent.click(screen.getByText('1.5x'));
		expect(props.setPlaybackRate).toHaveBeenCalledWith(1.5);

		unmount();
		clearTimers();
	});

	it('auto-hides controls when playing and timer expires', () => {
		jest.useFakeTimers();
		const props = {
			isPlaying: true,
			currentTime: 3661,
			duration: 7200,
			volume: 0,
			playbackRate: 1,
			isFullscreen: false,
			seekTo: jest.fn(),
			setVolume: jest.fn(),
			setPlaybackRate: jest.fn(),
			toggleFullscreen: jest.fn(),
			togglePlayPause: jest.fn(),
			nextVideo: jest.fn(),
			previousVideo: jest.fn(),
		};
		const { container } = render(<VideoControls {...props} />);
		expect(screen.getByText('1:01:01')).toBeInTheDocument();

		const volumeButton = container.querySelectorAll('.right-controls button')[0] as HTMLButtonElement;
		fireEvent.click(volumeButton);
		expect(props.setVolume).toHaveBeenCalledWith(0.5);

		act(() => {
			jest.advanceTimersByTime(3100);
		});
		expect(container.querySelector('.video-controls')).toBeNull();
		jest.useRealTimers();
	});

	it('keeps controls visible when not playing', () => {
		jest.useFakeTimers();
		const props = {
			isPlaying: false,
			currentTime: 0,
			duration: 100,
			volume: 1,
			playbackRate: 1,
			isFullscreen: true,
			seekTo: jest.fn(),
			setVolume: jest.fn(),
			setPlaybackRate: jest.fn(),
			toggleFullscreen: jest.fn(),
			togglePlayPause: jest.fn(),
			nextVideo: jest.fn(),
			previousVideo: jest.fn(),
		};
		const { container } = render(<VideoControls {...props} />);
		act(() => {
			jest.advanceTimersByTime(3500);
		});
		expect(container.querySelector('.video-controls')).toBeTruthy();
		jest.useRealTimers();
	});

	it('renders settings and applies changes', () => {
		const onClose = jest.fn();
		const setPlaybackRate = jest.fn();
		const setQuality = jest.fn();
		render(
			<VideoSettings
				anchorEl={document.body}
				onClose={onClose}
				playbackRate={1}
				setPlaybackRate={setPlaybackRate}
				quality='auto'
				setQuality={setQuality}
			/>,
		);
		fireEvent.click(screen.getByText('1.25x'));
		expect(setPlaybackRate).toHaveBeenCalledWith(1.25);
		fireEvent.click(screen.getByText('1080p'));
		expect(setQuality).toHaveBeenCalledWith('1080p');
		expect(onClose).toHaveBeenCalled();
	});

	it('renders progress bar and seeks', () => {
		const seekTo = jest.fn();
		render(<VideoProgressBar currentTime={10} duration={100} seekTo={seekTo} />);
		const bar = document.querySelector('.video-progress-bar') as HTMLElement;
		Object.defineProperty(bar, 'getBoundingClientRect', {
			value: () => ({ left: 0, width: 100 }),
		});
		fireEvent.mouseMove(bar, { clientX: 50 });
		fireEvent.click(bar, { clientX: 20 });
		fireEvent.mouseDown(bar);
		fireEvent.mouseMove(document, { clientX: 80 });
		fireEvent.mouseUp(document);
		fireEvent.mouseLeave(bar);
		expect(seekTo).toHaveBeenCalled();
	});

	it('handles zero duration and bounded seek math', () => {
		const seekTo = jest.fn();
		render(<VideoProgressBar currentTime={0} duration={0} seekTo={seekTo} />);
		const bar = document.querySelector('.video-progress-bar') as HTMLElement;
		Object.defineProperty(bar, 'getBoundingClientRect', {
			value: () => ({ left: 10, width: 100 }),
		});
		fireEvent.mouseMove(bar, { clientX: -50 });
		fireEvent.click(bar, { clientX: 9999 });
		fireEvent.mouseDown(bar);
		fireEvent.mouseMove(document, { clientX: -999 });
		fireEvent.mouseUp(document);
		expect(seekTo).toHaveBeenCalledWith(0);
	});

	it('formats invalid times in progress bar', () => {
		render(<VideoProgressBar currentTime={Number.NaN} duration={Number.NaN} seekTo={jest.fn()} />);
		expect(screen.getAllByText('0:00').length).toBeGreaterThan(0);
	});

	it('renders video thumbnail metadata and triggers play', () => {
		const onPlay = jest.fn();
		const video: any = {
			id: 1,
			name: 'sample.mp4',
			size: 1024,
			metadata: { duration: '01:00', width: 1920, height: 1080, codec_name: 'h264', format_name: 'Sample' },
		};
		render(<VideoThumbnail video={video} onPlay={onPlay} />);
		expect(screen.getByText('Sample')).toBeInTheDocument();
		fireEvent.click(screen.getByRole('button', { name: /Play Sample/i }));
		expect(onPlay).toHaveBeenCalled();
		fireEvent.keyDown(screen.getByRole('button', { name: /Play Sample/i }), { key: 'Enter' });
		expect(onPlay).toHaveBeenCalledTimes(2);
	});

	it('binds video player events and metadata display', () => {
		const ref = createRef<HTMLVideoElement>();
		const onVideoEnded = jest.fn();
		const setCurrentTime = jest.fn();
		const setDuration = jest.fn();
		const onBack = jest.fn();
		const currentVideo: any = {
			name: 'v.mp4',
		};
		const { container } = render(
			<VideoPlayer
				currentVideo={currentVideo}
				videoRef={ref}
				setCurrentTime={setCurrentTime}
				setDuration={setDuration}
				onBack={onBack}
				onVideoEnded={onVideoEnded}
				originBadgeLabel='Series'
				contextDescription='From My Show'
				metadataLine='My Show • Item 1 of 3'
			/>,
		);
		const video = container.querySelector('video') as HTMLVideoElement;
		Object.defineProperty(video, 'currentTime', { value: 12, configurable: true });
		Object.defineProperty(video, 'duration', { value: 30, configurable: true });
		expect(screen.getByText('v.mp4')).toBeInTheDocument();
		expect(screen.getByText('From My Show')).toBeInTheDocument();
		expect(screen.getByText('My Show • Item 1 of 3')).toBeInTheDocument();
		fireEvent(video, new Event('timeupdate'));
		fireEvent(video, new Event('loadedmetadata'));
		fireEvent(video, new Event('ended'));
		expect(onVideoEnded).toHaveBeenCalled();
		fireEvent.click(screen.getByText('Voltar'));
		expect(onBack).toHaveBeenCalled();
	});

	it('renders player/video thumbnail fallback metadata branches', () => {
		const onPlay = jest.fn();
		render(<VideoThumbnail video={{ id: 2, name: 'raw-video.mkv', size: 0 } as any} onPlay={onPlay} />);
		expect(screen.getByText('raw-video.mkv')).toBeInTheDocument();
		expect(screen.getAllByText('Unknown').length).toBeGreaterThan(0);
		expect(screen.getByText('0 B')).toBeInTheDocument();
		fireEvent.keyDown(screen.getByRole('button', { name: /Play raw-video.mkv/i }), { key: ' ' });
		expect(onPlay).toHaveBeenCalled();

		const ref = createRef<HTMLVideoElement>();
		render(
			<VideoPlayer
				currentVideo={null}
				videoRef={ref}
				setCurrentTime={jest.fn()}
				setDuration={jest.fn()}
				onBack={jest.fn()}
				onVideoEnded={jest.fn()}
				originBadgeLabel='Videos'
				contextDescription='From Videos'
				metadataLine=''
			/>,
		);
		expect(screen.getByText('No video playing')).toBeInTheDocument();
	});

	it('handles player with inert ref object and partial metadata', () => {
		const inertRef: any = {
			get current() {
				return null;
			},
			set current(_value: HTMLVideoElement | null) {},
		};
		render(
			<VideoPlayer
				currentVideo={{ name: 'just-name.mp4' } as any}
				videoRef={inertRef}
				setCurrentTime={jest.fn()}
				setDuration={jest.fn()}
				onBack={jest.fn()}
				onVideoEnded={jest.fn()}
				originBadgeLabel='Videos'
				contextDescription='From Files'
				metadataLine='Collection • Item 1 of 1'
			/>,
		);
		expect(screen.getByText('just-name.mp4')).toBeInTheDocument();
		expect(screen.getByText('Collection • Item 1 of 1')).toBeInTheDocument();
	});
});
