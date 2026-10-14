import { act, fireEvent, render, screen } from '@testing-library/react';
import VideoPlayerScreen from './VideoPlayerScreen';

const mockUseVideoPlayerScreen = jest.fn();
const mockOpenVideo = jest.fn();

const buildScreenState = () => ({
    isInvalidVideoId: false,
    handleBack: jest.fn(),
    handlePlaybackEnded: jest.fn(),
    upNext: {
        isOpen: false,
        isCounting: false,
        secondsRemaining: 8,
        isAutoplayEnabled: true,
        open: jest.fn(),
        cancel: jest.fn(),
        playNow: jest.fn(),
        setAutoplayEnabled: jest.fn(),
    },
    openVideo: mockOpenVideo,
    currentVideo: { id: 30, name: 'Episode 1' },
    contextTitle: 'My Show',
    originBadgeLabel: 'Series',
    contextDescription: 'From My Show',
    metadataLine: 'My Show • Item 1 of 3',
    nextItem: {
        id: 131,
        displayTitle: 'Episode 2',
        sequenceLabel: 'S01E02',
        status: 'not_started',
        progress_pct: 25,
        video: { id: 31, name: 'Episode 2' },
    },
    relatedItems: [
        {
            id: 132,
            displayTitle: 'Episode 3',
            sequenceLabel: 'S01E03',
            status: 'completed',
            progress_pct: 100,
            video: { id: 32, name: 'Episode 3' },
        },
    ],
    relatedTitle: 'Next episodes',
    hasNextVideo: true,
    hasPreviousVideo: false,
    videoRef: { current: null },
    registerPlayerContainer: jest.fn(),
    fullscreenPortalContainer: null,
    seekTo: jest.fn(),
    setVolume: jest.fn(),
    toggleMute: jest.fn(),
    setPlaybackRate: jest.fn(),
    toggleFullscreen: jest.fn(),
    togglePlayPause: jest.fn(),
    nextVideo: jest.fn(),
    previousVideo: jest.fn(),
    status: 'playing',
    currentTime: 12,
    duration: 120,
    volume: 0.5,
    isMuted: false,
    playbackRate: 1,
    isFullscreen: false,
    setCurrentTime: jest.fn(),
    setDuration: jest.fn(),
    subtitles: [],
    selectedSubtitleKey: null,
    selectSubtitle: jest.fn(),
    toggleSubtitles: jest.fn<'enabled' | 'disabled' | 'unavailable', []>(() => 'unavailable'),
});

jest.mock('./useVideoPlayerScreen', () => ({
    __esModule: true,
    default: () => mockUseVideoPlayerScreen(),
}));

jest.mock('@/features/videos/videoControls/videoControls', () => ({
    __esModule: true,
    default: (props: any) => (
        <div data-testid="video-controls">
            {String(props.canGoPrevious)}-{String(props.canGoNext)}
            <button type="button" onClick={props.onOpenShortcutsHelp}>
                open-shortcuts-help
            </button>
        </div>
    ),
}));

jest.mock('@/features/videos/videoPlayer/videoPlayer', () => ({
    __esModule: true,
    default: (props: any) => (
        <div data-testid="video-player">
            <div>{props.currentVideo?.name ?? 'no-video'}</div>
            <div>{props.originBadgeLabel}</div>
            <div>{props.contextDescription}</div>
            <div>{props.metadataLine}</div>
            <button type="button" onClick={props.onOpenDetails}>
                open-details
            </button>
            {props.children}
        </div>
    ),
}));

jest.mock('@/features/videos/videoDetails/VideoDetailsPanel', () => ({
    __esModule: true,
    default: (props: any) =>
        props.video ? (
            <div data-testid="details-panel">
                {props.video.name}
                <button type="button" onClick={props.onClose}>
                    close-details
                </button>
            </div>
        ) : null,
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string>) => {
            const translations: Record<string, string> = {
                VIDEO_INVALID_ID: 'Invalid video ID',
                VIDEO_PLAYER_UP_NEXT: 'Up next',
                VIDEO_STATUS_NOT_STARTED: 'Not started',
                VIDEO_STATUS_IN_PROGRESS: 'In progress',
                VIDEO_STATUS_COMPLETED: 'Completed',
                VIDEO_PLAYER_OPEN_VIDEO: 'Open {{name}}',
            };
            const template = translations[key] ?? key;
            return Object.entries(params ?? {}).reduce(
                (result, [name, value]) => result.replace(`{{${name}}}`, value),
                template
            );
        },
    }),
}));

describe('features/videos/videoPlayer/VideoPlayerScreen', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockUseVideoPlayerScreen.mockReturnValue(buildScreenState());
    });

    it('renders the invalid state when the route id is missing', () => {
        mockUseVideoPlayerScreen.mockReturnValueOnce({
            isInvalidVideoId: true,
        });

        render(<VideoPlayerScreen />);

        expect(screen.getByText('Invalid video ID')).toBeInTheDocument();
    });

    it('opens and closes the details panel for the current video', () => {
        render(<VideoPlayerScreen />);

        expect(screen.queryByTestId('details-panel')).toBeNull();

        fireEvent.click(screen.getByText('open-details'));
        expect(screen.getByTestId('details-panel')).toHaveTextContent('Episode 1');

        fireEvent.click(screen.getByText('close-details'));
        expect(screen.queryByTestId('details-panel')).toBeNull();
    });

    it('renders context actions and opens suggested videos from the panel', () => {
        render(<VideoPlayerScreen />);

        expect(screen.getByTestId('video-player')).toHaveTextContent('Episode 1');
        expect(screen.getByTestId('video-controls')).toHaveTextContent('false-true');
        expect(screen.getByText('My Show')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Open Episode 2' })).toBeInTheDocument();
        expect(screen.getByText('Next episodes')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'Open Episode 2' }));
        expect(mockOpenVideo).toHaveBeenCalledWith(31);

        fireEvent.click(screen.getByRole('button', { name: 'Open Episode 3' }));
        expect(mockOpenVideo).toHaveBeenCalledWith(32);
    });

    it('hides optional suggestion sections when the player has no contextual queue', () => {
        mockUseVideoPlayerScreen.mockReturnValueOnce({
            ...buildScreenState(),
            isInvalidVideoId: false,
            nextItem: null,
            relatedItems: [],
            hasNextVideo: false,
            hasPreviousVideo: false,
        });

        render(<VideoPlayerScreen />);

        expect(screen.queryByText('Up next')).not.toBeInTheDocument();
        expect(screen.queryByText('Next episodes')).not.toBeInTheDocument();
        expect(screen.getByTestId('video-controls')).toHaveTextContent('false-false');
    });
    describe('keyboard shortcuts', () => {
        const pressKey = (init: KeyboardEventInit, target: EventTarget = document.body) => {
            const event = new KeyboardEvent('keydown', {
                bubbles: true,
                cancelable: true,
                ...init,
            });
            act(() => {
                target.dispatchEvent(event);
            });
            return event;
        };

        it('toggles playback with Space and prevents page scrolling', () => {
            const state = buildScreenState();
            mockUseVideoPlayerScreen.mockReturnValue(state);
            render(<VideoPlayerScreen />);

            const event = pressKey({ key: ' ' });

            expect(event.defaultPrevented).toBe(true);
            expect(state.togglePlayPause).toHaveBeenCalledTimes(1);
            expect(screen.getByRole('status')).toHaveTextContent('VIDEO_SHORTCUT_FEEDBACK_PAUSE');
        });

        it('seeks with the arrow keys within the duration', () => {
            const state = { ...buildScreenState(), currentTime: 118 };
            mockUseVideoPlayerScreen.mockReturnValue(state);
            render(<VideoPlayerScreen />);

            pressKey({ key: 'ArrowRight' });
            expect(state.seekTo).toHaveBeenLastCalledWith(120);

            pressKey({ key: 'Home' });
            expect(state.seekTo).toHaveBeenLastCalledWith(0);

            pressKey({ key: '5' });
            expect(state.seekTo).toHaveBeenLastCalledWith(60);
        });

        it('adjusts volume, speed and fullscreen', () => {
            const state = buildScreenState();
            mockUseVideoPlayerScreen.mockReturnValue(state);
            render(<VideoPlayerScreen />);

            pressKey({ key: 'ArrowUp' });
            expect(state.setVolume).toHaveBeenLastCalledWith(0.55);
            pressKey({ key: '>', shiftKey: true });
            expect(state.setPlaybackRate).toHaveBeenLastCalledWith(1.25);
            pressKey({ key: 'f' });
            expect(state.toggleFullscreen).toHaveBeenCalledTimes(1);
        });

        it('toggles subtitles with C and shows the feedback', () => {
            const state = buildScreenState();
            state.toggleSubtitles.mockReturnValue('enabled');
            mockUseVideoPlayerScreen.mockReturnValue(state);
            render(<VideoPlayerScreen />);

            pressKey({ key: 'c' });

            expect(state.toggleSubtitles).toHaveBeenCalledTimes(1);
            expect(screen.getByRole('status')).toHaveTextContent(
                'VIDEO_SHORTCUT_FEEDBACK_SUBTITLES_ON'
            );
        });

        it('renders the up next card inside the player when it is open', () => {
            const state = buildScreenState();
            state.upNext = { ...state.upNext, isOpen: true, isCounting: true };
            mockUseVideoPlayerScreen.mockReturnValue(state);
            render(<VideoPlayerScreen />);

            const player = screen.getByTestId('video-player');
            expect(player).toContainElement(screen.getByText('VIDEO_UP_NEXT_PLAY_NOW'));

            fireEvent.click(screen.getByText('VIDEO_UP_NEXT_PLAY_NOW'));
            fireEvent.click(screen.getByText('VIDEO_UP_NEXT_CANCEL'));
            expect(state.upNext.playNow).toHaveBeenCalledTimes(1);
            expect(state.upNext.cancel).toHaveBeenCalledTimes(1);
        });

        it('hides the up next card when closed or when there is no next item', () => {
            const closedState = buildScreenState();
            mockUseVideoPlayerScreen.mockReturnValue(closedState);
            const { unmount } = render(<VideoPlayerScreen />);
            expect(screen.queryByText('VIDEO_UP_NEXT_PLAY_NOW')).not.toBeInTheDocument();
            unmount();

            const withoutNext = buildScreenState();
            withoutNext.upNext = { ...withoutNext.upNext, isOpen: true };
            mockUseVideoPlayerScreen.mockReturnValue({ ...withoutNext, nextItem: null });
            render(<VideoPlayerScreen />);
            expect(screen.queryByText('VIDEO_UP_NEXT_PLAY_NOW')).not.toBeInTheDocument();
        });

        it('navigates the playlist only where available', () => {
            const state = buildScreenState();
            mockUseVideoPlayerScreen.mockReturnValue(state);
            render(<VideoPlayerScreen />);

            pressKey({ key: 'N', shiftKey: true });
            pressKey({ key: 'P', shiftKey: true });

            expect(state.nextVideo).toHaveBeenCalledTimes(1);
            expect(state.previousVideo).not.toHaveBeenCalled();
        });

        it('opens the help dialog with the question mark and from the controls button', () => {
            mockUseVideoPlayerScreen.mockReturnValue(buildScreenState());
            render(<VideoPlayerScreen />);

            pressKey({ key: '?', shiftKey: true });
            expect(screen.getByRole('dialog')).toBeInTheDocument();

            fireEvent.keyDown(screen.getByRole('dialog'), { key: 'Escape' });
        });

        it('opens the help dialog from the controls button', () => {
            mockUseVideoPlayerScreen.mockReturnValue(buildScreenState());
            render(<VideoPlayerScreen />);

            fireEvent.click(screen.getByText('open-shortcuts-help'));

            expect(screen.getByRole('dialog')).toBeInTheDocument();
        });

        it('ignores shortcuts while the help dialog is open', () => {
            const state = buildScreenState();
            mockUseVideoPlayerScreen.mockReturnValue(state);
            render(<VideoPlayerScreen />);
            fireEvent.click(screen.getByText('open-shortcuts-help'));

            pressKey({ key: ' ' });

            expect(state.togglePlayPause).not.toHaveBeenCalled();
        });

        it('ignores shortcuts while the details panel is open', () => {
            const state = buildScreenState();
            mockUseVideoPlayerScreen.mockReturnValue(state);
            render(<VideoPlayerScreen />);
            fireEvent.click(screen.getByText('open-details'));

            const event = pressKey({ key: ' ' });

            expect(event.defaultPrevented).toBe(false);
            expect(state.togglePlayPause).not.toHaveBeenCalled();
        });

        it('ignores shortcuts typed in inputs and stops listening after unmount', () => {
            const state = buildScreenState();
            mockUseVideoPlayerScreen.mockReturnValue(state);
            const { unmount } = render(<VideoPlayerScreen />);
            const input = document.createElement('input');
            document.body.appendChild(input);

            pressKey({ key: 'k' }, input);
            expect(state.togglePlayPause).not.toHaveBeenCalled();
            input.remove();

            unmount();
            pressKey({ key: 'k' });
            expect(state.togglePlayPause).not.toHaveBeenCalled();
        });
    });
});
