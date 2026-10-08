import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import GlobalPlayerControl from './GlobalPlayerControl';

const mockUseGlobalMusic = jest.fn();

jest.mock('@mui/material/useMediaQuery', () => ({
    __esModule: true,
    default: () => true,
}));
jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => mockUseGlobalMusic(),
}));
jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const baseApi = () => ({
    hasQueue: true,
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
    currentTrack: { name: 'Test Song', metadata: { title: 'Meta Title', artist: 'Meta Artist' } },
    playbackContext: undefined,
    toggleQueue: jest.fn(),
    queueOpen: false,
    setQueueOpen: jest.fn(),
    queue: [],
    currentIndex: undefined,
});

describe('GlobalPlayerControl (compact mobile mode)', () => {
    it('shows a progress bar reflecting playback position', () => {
        mockUseGlobalMusic.mockReturnValue(baseApi());
        render(<GlobalPlayerControl />);
        expect(screen.getByLabelText('PLAYER_ARIA_PROGRESS')).toHaveAttribute(
            'aria-valuenow',
            '25'
        );
    });

    it('shows zero progress when duration is unknown', () => {
        mockUseGlobalMusic.mockReturnValue({ ...baseApi(), duration: 0 });
        render(<GlobalPlayerControl />);
        expect(screen.getByLabelText('PLAYER_ARIA_PROGRESS')).toHaveAttribute('aria-valuenow', '0');
    });

    it('opens the expanded sheet when tapping the mini-player and closes it', async () => {
        mockUseGlobalMusic.mockReturnValue(baseApi());
        render(<GlobalPlayerControl />);
        expect(
            screen.queryByRole('button', { name: 'PLAYER_ARIA_CLOSE_EXPANDED' })
        ).not.toBeInTheDocument();

        fireEvent.click(screen.getAllByText('Meta Title')[0] as HTMLElement);
        fireEvent.click(await screen.findByRole('button', { name: 'PLAYER_ARIA_CLOSE_EXPANDED' }));
        await waitFor(() =>
            expect(
                screen.queryByRole('button', { name: 'PLAYER_ARIA_CLOSE_EXPANDED' })
            ).not.toBeInTheDocument()
        );
    });

    it('opens the expanded sheet from the keyboard', async () => {
        mockUseGlobalMusic.mockReturnValue(baseApi());
        render(<GlobalPlayerControl />);
        const openButton = screen.getByLabelText('PLAYER_ARIA_OPEN_EXPANDED');
        fireEvent.keyDown(openButton, { key: 'a' });
        expect(
            screen.queryByRole('button', { name: 'PLAYER_ARIA_CLOSE_EXPANDED' })
        ).not.toBeInTheDocument();
        fireEvent.keyDown(openButton, { key: 'Enter' });
        expect(
            await screen.findByRole('button', { name: 'PLAYER_ARIA_CLOSE_EXPANDED' })
        ).toBeInTheDocument();
    });

    it('does not open the sheet when tapping a mini-player button', () => {
        const api = baseApi();
        mockUseGlobalMusic.mockReturnValue(api);
        render(<GlobalPlayerControl />);
        fireEvent.click(screen.getByRole('button', { name: 'PLAYER_ARIA_NEXT' }));
        expect(api.next).toHaveBeenCalled();
        expect(
            screen.queryByRole('button', { name: 'PLAYER_ARIA_CLOSE_EXPANDED' })
        ).not.toBeInTheDocument();
    });
});
