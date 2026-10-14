import { act, fireEvent, screen } from '@testing-library/react';
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
    playbackError: null,
    reportPlaybackError: jest.fn(),
    retryPlayback: jest.fn(),
    ...overrides,
});

describe('VideoPlayerScreen fullscreen', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('mounts against an absent backend with the default state', () => {
        mockUseVideoPlayerScreen.mockReturnValue(buildScreenState());

        const { container } = renderWithoutBackend(<VideoPlayerScreen />);

        expect(container.querySelector('video')).not.toBeNull();
        expect(screen.getByRole('button', { name: 'VIDEO_FULLSCREEN' })).toBeInTheDocument();
    });

    it('registers the player section as the fullscreen container', () => {
        const state = buildScreenState();
        mockUseVideoPlayerScreen.mockReturnValue(state);

        renderWithoutBackend(<VideoPlayerScreen />);

        expect(state.registerPlayerContainer).toHaveBeenCalledWith(expect.any(HTMLElement));
        expect((state.registerPlayerContainer as jest.Mock).mock.calls[0][0].tagName).toBe(
            'SECTION'
        );
    });

    it('toggles fullscreen with the F key, the button and a double click on the video', () => {
        const state = buildScreenState();
        mockUseVideoPlayerScreen.mockReturnValue(state);
        const { container } = renderWithoutBackend(<VideoPlayerScreen />);

        fireEvent.keyDown(document.body, { key: 'f' });
        fireEvent.click(screen.getByRole('button', { name: 'VIDEO_FULLSCREEN' }));
        fireEvent.doubleClick(container.querySelector('video')!);

        expect(state.toggleFullscreen).toHaveBeenCalledTimes(3);
    });

    it('labels the button as exit while fullscreen', () => {
        mockUseVideoPlayerScreen.mockReturnValue(buildScreenState({ isFullscreen: true }));

        renderWithoutBackend(<VideoPlayerScreen />);

        expect(screen.getByRole('button', { name: 'VIDEO_EXIT_FULLSCREEN' })).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'VIDEO_FULLSCREEN' })).toBeNull();
    });

    describe('idle behavior', () => {
        beforeEach(() => {
            jest.useFakeTimers();
        });

        afterEach(() => {
            jest.useRealTimers();
        });

        it('hides the controls and the cursor after inactivity in fullscreen, and restores them on movement', () => {
            mockUseVideoPlayerScreen.mockReturnValue(buildScreenState({ isFullscreen: true }));
            const { container } = renderWithoutBackend(<VideoPlayerScreen />);
            const playerSection = container.querySelector('section')!;
            expect(playerSection.className).not.toContain('cursorHidden');

            act(() => {
                jest.advanceTimersByTime(3100);
            });
            expect(playerSection.className).toContain('cursorHidden');
            expect(container.querySelector('.video-controls')).toBeNull();

            act(() => {
                fireEvent.mouseMove(document);
            });
            expect(playerSection.className).not.toContain('cursorHidden');
            expect(container.querySelector('.video-controls')).not.toBeNull();
        });

        it('keeps the cursor outside fullscreen while the controls hide as before', () => {
            mockUseVideoPlayerScreen.mockReturnValue(buildScreenState({ isFullscreen: false }));
            const { container } = renderWithoutBackend(<VideoPlayerScreen />);

            act(() => {
                jest.advanceTimersByTime(3100);
            });

            expect(container.querySelector('section')!.className).not.toContain('cursorHidden');
            expect(container.querySelector('.video-controls')).toBeNull();
        });
    });

    describe('portals while fullscreen', () => {
        it('renders the shortcuts dialog inside the fullscreen container', () => {
            const fullscreenContainer = document.createElement('div');
            document.body.appendChild(fullscreenContainer);
            mockUseVideoPlayerScreen.mockReturnValue(
                buildScreenState({
                    isFullscreen: true,
                    fullscreenPortalContainer: fullscreenContainer,
                })
            );
            renderWithoutBackend(<VideoPlayerScreen />);

            fireEvent.click(screen.getByRole('button', { name: 'VIDEO_SHORTCUTS_HELP_BUTTON' }));

            expect(fullscreenContainer.querySelector('[role="dialog"]')).not.toBeNull();
            fullscreenContainer.remove();
        });

        it('renders the details drawer inside the fullscreen container', async () => {
            const fullscreenContainer = document.createElement('div');
            document.body.appendChild(fullscreenContainer);
            mockUseVideoPlayerScreen.mockReturnValue(
                buildScreenState({
                    isFullscreen: true,
                    fullscreenPortalContainer: fullscreenContainer,
                })
            );
            renderWithoutBackend(<VideoPlayerScreen />);

            fireEvent.click(screen.getByRole('button', { name: 'FILES_DETAILS' }));

            expect(
                await screen.findByRole('region', { name: 'VIDEO_DETAILS_TITLE' })
            ).toBeInTheDocument();
            expect(fullscreenContainer.querySelector('.MuiDrawer-root')).not.toBeNull();
            fullscreenContainer.remove();
        });
    });
});
