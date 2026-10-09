import { act, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import GlobalShortcutsProvider from '@/components/shortcuts/GlobalShortcutsProvider';
import { requestShortcutsHelp } from '@/components/layout/appCommandEvents';
import { usePlayerShortcuts } from './usePlayerShortcuts';

const mockUseGlobalMusic = jest.fn();

jest.mock('@/features/music/providers/GlobalMusicProvider', () => ({
    useGlobalMusic: () => mockUseGlobalMusic(),
}));

const PlayerShortcutsProbe = () => {
    usePlayerShortcuts();
    return null;
};

describe('usePlayerShortcuts', () => {
    const buildPlayer = (overrides: Record<string, unknown> = {}) => ({
        hasQueue: true,
        currentTime: 30,
        duration: 100,
        volume: 0.5,
        repeatMode: 'none',
        togglePlayPause: jest.fn(),
        next: jest.fn(),
        previous: jest.fn(),
        seek: jest.fn(),
        setVolume: jest.fn(),
        toggleShuffle: jest.fn(),
        setRepeatMode: jest.fn(),
        ...overrides,
    });

    const press = (
        key: string,
        init: KeyboardEventInit = {},
        target: Element | Document = document.body
    ) => {
        let wasNotPrevented = true;
        act(() => {
            wasNotPrevented = fireEvent.keyDown(target, { key, ...init });
        });
        return wasNotPrevented;
    };

    const renderWithPlayer = (
        player: ReturnType<typeof buildPlayer>,
        extra: React.ReactNode = null
    ) => {
        mockUseGlobalMusic.mockReturnValue(player);
        return render(
            <MemoryRouter>
                <GlobalShortcutsProvider>
                    <PlayerShortcutsProbe />
                    <input aria-label="field" />
                    {extra}
                </GlobalShortcutsProvider>
            </MemoryRouter>
        );
    };

    beforeEach(() => mockUseGlobalMusic.mockReset());

    it('calls the matching player action for each key', () => {
        const player = buildPlayer();
        renderWithPlayer(player);

        press(' ');
        press('ArrowRight', { shiftKey: true });
        press('ArrowLeft', { shiftKey: true });
        press('ArrowRight');
        press('ArrowLeft');
        press('ArrowUp');
        press('ArrowDown');
        press('m');
        press('s');
        press('r');

        expect(player.togglePlayPause).toHaveBeenCalledTimes(1);
        expect(player.next).toHaveBeenCalledTimes(1);
        expect(player.previous).toHaveBeenCalledTimes(1);
        expect(player.seek).toHaveBeenNthCalledWith(1, 40);
        expect(player.seek).toHaveBeenNthCalledWith(2, 20);
        expect(player.setVolume).toHaveBeenNthCalledWith(1, 0.55);
        expect(player.setVolume).toHaveBeenNthCalledWith(2, 0.45);
        expect(player.setVolume).toHaveBeenNthCalledWith(3, 0);
        expect(player.toggleShuffle).toHaveBeenCalledTimes(1);
        expect(player.setRepeatMode).toHaveBeenCalledWith('all');
    });

    it('unmutes to the default volume and clamps seek and volume', () => {
        const player = buildPlayer({ currentTime: 95, volume: 0 });
        renderWithPlayer(player);

        press('m');
        press('ArrowRight');
        press('ArrowDown');

        expect(player.setVolume).toHaveBeenNthCalledWith(1, 0.7);
        expect(player.seek).toHaveBeenCalledWith(100);
        expect(player.setVolume).toHaveBeenNthCalledWith(2, 0);
    });

    it('prevents the page from scrolling when handling space', () => {
        renderWithPlayer(buildPlayer());

        expect(press(' ')).toBe(false);
    });

    it('does nothing without a queue', () => {
        const player = buildPlayer({ hasQueue: false });
        renderWithPlayer(player);

        press(' ');
        press('m');

        expect(player.togglePlayPause).not.toHaveBeenCalled();
        expect(player.setVolume).not.toHaveBeenCalled();
    });

    it('ignores keys typed into an input', () => {
        const player = buildPlayer();
        renderWithPlayer(player);

        press(' ', {}, screen.getByLabelText('field'));

        expect(player.togglePlayPause).not.toHaveBeenCalled();
    });

    it('ignores space on a focused button so the button keeps its own activation', () => {
        const player = buildPlayer();
        renderWithPlayer(player, <button type="button">press</button>);

        press(' ', {}, screen.getByRole('button', { name: 'press' }));

        expect(player.togglePlayPause).not.toHaveBeenCalled();
    });

    it('ignores keys while a modal is open', () => {
        const player = buildPlayer();
        renderWithPlayer(player);
        const modalElement = document.createElement('div');
        modalElement.className = 'MuiModal-root';
        document.body.appendChild(modalElement);

        press(' ');
        modalElement.remove();

        expect(player.togglePlayPause).not.toHaveBeenCalled();
    });

    it.each(['ctrlKey', 'metaKey', 'altKey'])('ignores keys pressed with %s', (modifier) => {
        const player = buildPlayer();
        renderWithPlayer(player);

        press('m', { [modifier]: true });

        expect(player.setVolume).not.toHaveBeenCalled();
    });

    it('ignores events another handler already prevented', () => {
        const player = buildPlayer();
        renderWithPlayer(player);
        const preventingListener = (event: KeyboardEvent) => event.preventDefault();
        document.addEventListener('keydown', preventingListener);

        press('ArrowRight');
        document.removeEventListener('keydown', preventingListener);

        expect(player.seek).not.toHaveBeenCalled();
    });

    it('does not steal the key that completes a go-to shortcut', () => {
        const player = buildPlayer();
        renderWithPlayer(player);

        press('g');
        press('s');
        press('s');

        expect(player.toggleShuffle).toHaveBeenCalledTimes(1);
    });

    it('lists the player shortcuts in the help dialog under the Player section', async () => {
        renderWithPlayer(buildPlayer());

        act(() => requestShortcutsHelp());

        expect(await screen.findByText('SHORTCUTS_SECTION_PLAYER')).toBeInTheDocument();
        expect(screen.getByText('PLAYER_SHORTCUT_PLAY_PAUSE')).toBeInTheDocument();
        expect(screen.getByText('PLAYER_SHORTCUT_REPEAT')).toBeInTheDocument();
    });

    it('omits the Player section from the help dialog without a queue', async () => {
        renderWithPlayer(buildPlayer({ hasQueue: false }));

        act(() => requestShortcutsHelp());

        await screen.findByText('SHORTCUTS_DIALOG_TITLE');
        expect(screen.queryByText('SHORTCUTS_SECTION_PLAYER')).not.toBeInTheDocument();
    });
});
