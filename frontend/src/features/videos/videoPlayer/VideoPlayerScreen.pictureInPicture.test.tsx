import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import VideoPlayerScreen from './VideoPlayerScreen';

const mockUseVideoPlayerScreen = jest.fn();

jest.mock('./useVideoPlayerScreen', () => ({
    __esModule: true,
    default: () => mockUseVideoPlayerScreen(),
}));

const buildScreenState = (overrides: Record<string, unknown> = {}) => ({
    isInvalidVideoId: false,
    handleBack: jest.fn(),
    handlePlaybackEnded: jest.fn(),
    openVideo: jest.fn(),
    currentVideo: { id: 30, name: 'Episode 1' },
    contextTitle: 'My Show',
    originBadgeLabel: 'Series',
    contextDescription: 'From My Show',
    metadataLine: '',
    nextItem: null,
    relatedItems: [],
    relatedTitle: '',
    hasNextVideo: true,
    hasPreviousVideo: false,
    videoRef: { current: null },
    registerPlayerContainer: jest.fn(),
    fullscreenPortalContainer: null,
    seekTo: jest.fn(),
    setVolume: jest.fn(),
    setPlaybackRate: jest.fn(),
    toggleFullscreen: jest.fn(),
    isPictureInPictureSupported: false,
    isPictureInPicture: false,
    togglePictureInPicture: jest.fn(),
    pictureInPictureError: null,
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
    playbackError: null,
    reportPlaybackError: jest.fn(),
    retryPlayback: jest.fn(),
    ...overrides,
});

describe('VideoPlayerScreen picture in picture', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('mounts against an absent backend without the picture in picture button', () => {
        mockUseVideoPlayerScreen.mockReturnValue(buildScreenState());

        renderWithoutBackend(<VideoPlayerScreen />);

        expect(screen.queryByRole('button', { name: 'VIDEO_PICTURE_IN_PICTURE' })).toBeNull();
    });

    it('toggles picture in picture with the button and the I key', () => {
        const state = buildScreenState({ isPictureInPictureSupported: true });
        mockUseVideoPlayerScreen.mockReturnValue(state);
        renderWithoutBackend(<VideoPlayerScreen />);

        fireEvent.click(screen.getByRole('button', { name: 'VIDEO_PICTURE_IN_PICTURE' }));
        fireEvent.keyDown(document.body, { key: 'i' });

        expect(state.togglePictureInPicture).toHaveBeenCalledTimes(2);
    });

    it('shows the request error as discreet feedback', () => {
        mockUseVideoPlayerScreen.mockReturnValue(
            buildScreenState({
                isPictureInPictureSupported: true,
                pictureInPictureError: { id: 1, messageKey: 'VIDEO_PICTURE_IN_PICTURE_ERROR' },
            })
        );

        renderWithoutBackend(<VideoPlayerScreen />);

        expect(screen.getByRole('status')).toHaveTextContent('VIDEO_PICTURE_IN_PICTURE_ERROR');
    });
});
