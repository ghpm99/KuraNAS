import { fireEvent, render, screen } from '@testing-library/react';
import VideoDetailsPanel from '@/features/videos/videoDetails/VideoDetailsPanel';
import PlayerShortcutsHelpDialog from '@/features/videos/playerShortcuts/PlayerShortcutsHelpDialog';
import VideoControls from '@/features/videos/videoControls/videoControls';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { PlayerPortalContainerContext, usePlayerPortalContainer } from './playerPortalContainer';

const PortalContainerProbe = () => {
    const portalContainer = usePlayerPortalContainer();
    return <span>{portalContainer ? 'has-container' : 'no-container'}</span>;
};

describe('videoPlayer/playerPortalContainer', () => {
    it('falls back to no container without a provider', () => {
        render(<PortalContainerProbe />);

        expect(screen.getByText('no-container')).toBeInTheDocument();
    });

    it('renders the shortcuts dialog and the details drawer in the document body by default', () => {
        renderWithoutBackend(
            <>
                <PlayerShortcutsHelpDialog isOpen />
                <VideoDetailsPanel video={{ id: 9 } as any} onClose={jest.fn()} />
            </>
        );

        expect(document.body.querySelector(':scope > .MuiDialog-root')).not.toBeNull();
        expect(document.body.querySelector(':scope > .MuiDrawer-root')).not.toBeNull();
    });

    describe('with a fullscreen container', () => {
        let fullscreenContainer: HTMLElement;

        beforeEach(() => {
            fullscreenContainer = document.createElement('div');
            document.body.appendChild(fullscreenContainer);
        });

        afterEach(() => {
            fullscreenContainer.remove();
        });

        it('renders the details drawer inside the container', async () => {
            renderWithoutBackend(
                <PlayerPortalContainerContext.Provider value={fullscreenContainer}>
                    <VideoDetailsPanel video={{ id: 9 } as any} onClose={jest.fn()} />
                </PlayerPortalContainerContext.Provider>
            );

            expect(await screen.findByRole('region', { name: 'VIDEO_DETAILS_TITLE' })).toBeTruthy();
            expect(fullscreenContainer.querySelector('.MuiDrawer-root')).not.toBeNull();
            expect(document.body.querySelector(':scope > .MuiDrawer-root')).toBeNull();
        });

        it('renders the shortcuts dialog inside the container', () => {
            render(
                <PlayerPortalContainerContext.Provider value={fullscreenContainer}>
                    <PlayerShortcutsHelpDialog isOpen />
                </PlayerPortalContainerContext.Provider>
            );

            expect(fullscreenContainer.querySelector('[role="dialog"]')).not.toBeNull();
            expect(document.body.querySelector(':scope > .MuiDialog-root')).toBeNull();
        });

        it('renders the playback speed menu and the tooltips inside the container', async () => {
            const { container } = render(
                <PlayerPortalContainerContext.Provider value={fullscreenContainer}>
                    <VideoControls
                        isPlaying={false}
                        currentTime={0}
                        duration={10}
                        volume={1}
                        playbackRate={1}
                        isFullscreen
                        seekTo={jest.fn()}
                        setVolume={jest.fn()}
                        setPlaybackRate={jest.fn()}
                        toggleFullscreen={jest.fn()}
                        togglePlayPause={jest.fn()}
                        nextVideo={jest.fn()}
                        previousVideo={jest.fn()}
                        onOpenShortcutsHelp={jest.fn()}
                    />
                </PlayerPortalContainerContext.Provider>
            );

            const settingsButton = container.querySelectorAll('.right-controls button')[1]!;
            fireEvent.click(settingsButton);
            expect(await screen.findByText('VIDEO_PLAYBACK_SPEED')).toBeInTheDocument();
            expect(fullscreenContainer.querySelector('.MuiMenu-root')).not.toBeNull();
            expect(document.body.querySelector(':scope > .MuiMenu-root')).toBeNull();
        });
    });
});
