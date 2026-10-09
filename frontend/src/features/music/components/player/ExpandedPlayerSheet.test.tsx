import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen } from '@testing-library/react';
import type { ComponentProps } from 'react';
import ExpandedPlayerSheet from './ExpandedPlayerSheet';
import { supportsProgrammaticVolume } from './supportsProgrammaticVolume';

const mockUseGlobalMusic = jest.fn();

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => mockUseGlobalMusic(),
}));
jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));
jest.mock('./supportsProgrammaticVolume', () => ({
    supportsProgrammaticVolume: jest.fn(() => true),
}));

const buildPlayer = () => ({
    isPlaying: false,
    currentTime: 30,
    duration: 120,
    volume: 0.5,
    shuffle: false,
    repeatMode: 'none',
    togglePlayPause: jest.fn(),
    next: jest.fn(),
    previous: jest.fn(),
    seek: jest.fn(),
    setVolume: jest.fn(),
    toggleShuffle: jest.fn(),
    setRepeatMode: jest.fn(),
    setQueueOpen: jest.fn(),
    currentTrack: { name: 'Song', metadata: { title: 'Song Title', artist: 'Song Artist' } },
});

const renderSheet = (
    onClose = jest.fn(),
    props: Partial<ComponentProps<typeof ExpandedPlayerSheet>> = {}
) =>
    render(
        <QueryClientProvider
            client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
        >
            <ExpandedPlayerSheet isOpen onOpen={jest.fn()} onClose={onClose} {...props} />
        </QueryClientProvider>
    );

describe('ExpandedPlayerSheet', () => {
    afterEach(() => {
        document.documentElement.removeAttribute('data-app-motion');
    });

    it('renders without a current track or finite timings', () => {
        mockUseGlobalMusic.mockReturnValue({
            ...buildPlayer(),
            currentTrack: undefined,
            currentTime: NaN,
            duration: NaN,
        });
        renderSheet();
        expect(screen.getByLabelText('PLAYER_ARIA_SEEK')).toBeInTheDocument();
    });

    it('shows track info and times', () => {
        mockUseGlobalMusic.mockReturnValue({ ...buildPlayer(), isPlaying: true });
        renderSheet();
        expect(screen.getByText('Song Title')).toBeInTheDocument();
        expect(screen.getByText('Song Artist')).toBeInTheDocument();
        expect(screen.getByText('0:30')).toBeInTheDocument();
        expect(screen.getByText('2:00')).toBeInTheDocument();
        expect(screen.getByLabelText('PLAYER_ARIA_PAUSE')).toBeInTheDocument();
    });

    it('calls the player context for every control', () => {
        const player = buildPlayer();
        mockUseGlobalMusic.mockReturnValue(player);
        renderSheet();

        fireEvent.click(screen.getByLabelText('PLAYER_ARIA_PLAY'));
        fireEvent.click(screen.getByLabelText('PLAYER_ARIA_NEXT'));
        fireEvent.click(screen.getByLabelText('PLAYER_ARIA_PREVIOUS'));
        fireEvent.click(screen.getByLabelText('PLAYER_ARIA_SHUFFLE'));
        fireEvent.click(screen.getByLabelText('PLAYER_ARIA_REPEAT'));
        fireEvent.click(screen.getByLabelText('PLAYER_ARIA_MUTE'));

        expect(player.togglePlayPause).toHaveBeenCalled();
        expect(player.next).toHaveBeenCalled();
        expect(player.previous).toHaveBeenCalled();
        expect(player.toggleShuffle).toHaveBeenCalled();
        expect(player.setRepeatMode).toHaveBeenCalledWith('all');
        expect(player.setVolume).toHaveBeenCalledWith(0);
    });

    it('seeks and changes volume through the sliders', () => {
        const player = buildPlayer();
        mockUseGlobalMusic.mockReturnValue(player);
        renderSheet();

        const seekSlider = screen.getByRole('slider', { name: 'PLAYER_ARIA_SEEK' });
        fireEvent.change(seekSlider, { target: { value: 60 } });
        expect(player.seek).toHaveBeenCalledWith(60);

        const volumeSlider = screen.getByRole('slider', { name: 'PLAYER_ARIA_VOLUME' });
        fireEvent.change(volumeSlider, { target: { value: 0.2 } });
        expect(player.setVolume).toHaveBeenCalledWith(0.2);
    });

    it('keeps seek silent while dragging and seeks once on release', () => {
        const player = buildPlayer();
        mockUseGlobalMusic.mockReturnValue(player);
        renderSheet();

        const seekSlider = screen.getByRole('slider', { name: 'PLAYER_ARIA_SEEK' });
        const sliderRoot = seekSlider.closest('.MuiSlider-root') as HTMLElement;
        sliderRoot.getBoundingClientRect = () =>
            ({ left: 0, width: 100, top: 0, height: 10, right: 100, bottom: 10 }) as DOMRect;

        fireEvent.mouseDown(sliderRoot, { clientX: 50, clientY: 5 });
        fireEvent.mouseMove(document, { clientX: 75, clientY: 5, buttons: 1 });

        expect(screen.getByText('1:30')).toBeInTheDocument();
        expect(player.seek).not.toHaveBeenCalled();

        fireEvent.mouseUp(document, { clientX: 75, clientY: 5 });

        expect(player.seek).toHaveBeenCalledTimes(1);
        expect(player.seek).toHaveBeenCalledWith(90);
        expect(screen.getByText('0:30')).toBeInTheDocument();
    });

    it('unmutes to 0.7 and reflects active repeat one', () => {
        const player = { ...buildPlayer(), volume: 0, repeatMode: 'one', shuffle: true };
        mockUseGlobalMusic.mockReturnValue(player);
        renderSheet();
        fireEvent.click(screen.getByLabelText('PLAYER_ARIA_UNMUTE'));
        expect(player.setVolume).toHaveBeenCalledWith(0.7);
    });

    it('hides volume controls when the platform cannot set volume', () => {
        (supportsProgrammaticVolume as jest.Mock).mockReturnValueOnce(false);
        mockUseGlobalMusic.mockReturnValue(buildPlayer());
        renderSheet();
        expect(screen.queryByLabelText('PLAYER_ARIA_VOLUME')).not.toBeInTheDocument();
    });

    it('opens the queue and closes the sheet', () => {
        const player = buildPlayer();
        const onClose = jest.fn();
        mockUseGlobalMusic.mockReturnValue(player);
        renderSheet(onClose);
        fireEvent.click(screen.getByLabelText('PLAYER_ARIA_QUEUE'));
        expect(onClose).toHaveBeenCalled();
        expect(player.setQueueOpen).toHaveBeenCalledWith(true);
    });

    it('closes through the collapse button', () => {
        const onClose = jest.fn();
        mockUseGlobalMusic.mockReturnValue(buildPlayer());
        renderSheet(onClose);
        fireEvent.click(screen.getByLabelText('PLAYER_ARIA_CLOSE_EXPANDED'));
        expect(onClose).toHaveBeenCalled();
    });

    it('renders when reduced motion is requested', () => {
        document.documentElement.setAttribute('data-app-motion', 'reduced');
        mockUseGlobalMusic.mockReturnValue(buildPlayer());
        renderSheet();
        expect(screen.getByLabelText('PLAYER_ARIA_SEEK')).toBeInTheDocument();
    });

    it('toggles between artwork and lyrics through the lyrics button', () => {
        const onViewChange = jest.fn();
        mockUseGlobalMusic.mockReturnValue(buildPlayer());
        renderSheet(jest.fn(), { onViewChange });
        fireEvent.click(screen.getByLabelText('PLAYER_ARIA_LYRICS'));
        expect(onViewChange).toHaveBeenCalledWith('lyrics');
    });

    it('shows the lyrics panel and returns to artwork from the lyrics view', () => {
        const onViewChange = jest.fn();
        mockUseGlobalMusic.mockReturnValue({
            ...buildPlayer(),
            currentTrack: { id: 3, name: 'Song', metadata: { lyrics: 'first line' } },
        });
        renderSheet(jest.fn(), { view: 'lyrics', onViewChange });
        expect(screen.getByText('first line')).toBeInTheDocument();
        fireEvent.click(screen.getByLabelText('PLAYER_ARIA_LYRICS'));
        expect(onViewChange).toHaveBeenCalledWith('artwork');
    });

    it('ignores the lyrics button when no view handler is provided', () => {
        mockUseGlobalMusic.mockReturnValue(buildPlayer());
        renderSheet();
        expect(() => fireEvent.click(screen.getByLabelText('PLAYER_ARIA_LYRICS'))).not.toThrow();
    });
});
