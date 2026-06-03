import { fireEvent, render, screen } from '@testing-library/react';
import VideoPlayerScreen from './VideoPlayerScreen';

const mockUseVideoPlayerScreen = jest.fn();
const mockOpenVideo = jest.fn();

const buildScreenState = () => ({
    isInvalidVideoId: false,
    handleBack: jest.fn(),
    handlePlaybackEnded: jest.fn(),
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
    seekTo: jest.fn(),
    setVolume: jest.fn(),
    setPlaybackRate: jest.fn(),
    toggleFullscreen: jest.fn(),
    togglePlayPause: jest.fn(),
    nextVideo: jest.fn(),
    previousVideo: jest.fn(),
    status: 'playing',
    currentTime: 12,
    duration: 120,
    volume: 0.5,
    playbackRate: 1,
    isFullscreen: false,
    setCurrentTime: jest.fn(),
    setDuration: jest.fn(),
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
            {props.children}
        </div>
    ),
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

describe('components/videos/videoPlayer/VideoPlayerScreen', () => {
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
});
