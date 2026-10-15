import { act, fireEvent, render, screen } from '@testing-library/react';
import { createRef } from 'react';
import VideoControls from '@/features/videos/videoControls/videoControls';
import VideoPlayer from '@/features/videos/videoPlayer/videoPlayer';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string) => {
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
                VIDEO_NO_VIDEO_PLAYING: 'No video playing',
                VIDEO_BACK: 'Voltar',
                VIDEO_AUTO: 'Auto',
                VIDEO_NORMAL: 'Normal',
                VIDEO_FULLSCREEN: 'Fullscreen',
                VIDEO_EXIT_FULLSCREEN: 'Exit Fullscreen',
            };
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
            isMuted: false,
            playbackRate: 1,
            isFullscreen: false,
            seekTo: jest.fn(),
            setVolume: jest.fn(),
            toggleMute: jest.fn(),
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

    it('renders nothing when the controls are hidden', () => {
        const props = {
            isPlaying: true,
            currentTime: 3661,
            duration: 7200,
            volume: 0,
            isMuted: false,
            playbackRate: 1,
            isFullscreen: false,
            areControlsVisible: false,
            seekTo: jest.fn(),
            setVolume: jest.fn(),
            toggleMute: jest.fn(),
            setPlaybackRate: jest.fn(),
            toggleFullscreen: jest.fn(),
            togglePlayPause: jest.fn(),
            nextVideo: jest.fn(),
            previousVideo: jest.fn(),
        };
        const { container } = render(<VideoControls {...props} />);
        expect(container.querySelector('.video-controls')).toBeNull();
    });

    it('formats hours and toggles mute from the volume button', () => {
        const props = {
            isPlaying: true,
            currentTime: 3661,
            duration: 7200,
            volume: 0,
            isMuted: false,
            playbackRate: 1,
            isFullscreen: false,
            seekTo: jest.fn(),
            setVolume: jest.fn(),
            toggleMute: jest.fn(),
            setPlaybackRate: jest.fn(),
            toggleFullscreen: jest.fn(),
            togglePlayPause: jest.fn(),
            nextVideo: jest.fn(),
            previousVideo: jest.fn(),
        };
        const { container } = render(<VideoControls {...props} />);
        expect(screen.getByText('1:01:01')).toBeInTheDocument();

        fireEvent.click(container.querySelectorAll('.right-controls button')[0]!);
        expect(props.toggleMute).toHaveBeenCalledTimes(1);
    });

    it('keeps controls visible when not playing', () => {
        jest.useFakeTimers();
        const props = {
            isPlaying: false,
            currentTime: 0,
            duration: 100,
            volume: 1,
            isMuted: false,
            playbackRate: 1,
            isFullscreen: true,
            seekTo: jest.fn(),
            setVolume: jest.fn(),
            toggleMute: jest.fn(),
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
                originBadgeLabel="Series"
                contextDescription="From My Show"
                metadataLine="My Show • Item 1 of 3"
            />
        );
        const video = container.querySelector('video') as HTMLVideoElement;
        Object.defineProperty(video, 'currentTime', {
            value: 12,
            configurable: true,
        });
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

    it('renders the player placeholder when there is no current video', () => {
        const ref = createRef<HTMLVideoElement>();
        render(
            <VideoPlayer
                currentVideo={null}
                videoRef={ref}
                setCurrentTime={jest.fn()}
                setDuration={jest.fn()}
                onBack={jest.fn()}
                onVideoEnded={jest.fn()}
                originBadgeLabel="Videos"
                contextDescription="From Videos"
                metadataLine=""
            />
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
                originBadgeLabel="Videos"
                contextDescription="From Files"
                metadataLine="Collection • Item 1 of 1"
            />
        );
        expect(screen.getByText('just-name.mp4')).toBeInTheDocument();
        expect(screen.getByText('Collection • Item 1 of 1')).toBeInTheDocument();
    });
});
