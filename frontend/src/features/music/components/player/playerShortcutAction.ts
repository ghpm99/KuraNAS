export type PlayerShortcutAction =
    | 'togglePlayPause'
    | 'next'
    | 'previous'
    | 'seekForward'
    | 'seekBackward'
    | 'volumeUp'
    | 'volumeDown'
    | 'toggleMute'
    | 'toggleShuffle'
    | 'cycleRepeat';

type PlayerKeyboardEventKeys = Pick<
    KeyboardEvent,
    'key' | 'ctrlKey' | 'metaKey' | 'altKey' | 'shiftKey'
>;

const actionsByPlainKey: Record<string, PlayerShortcutAction> = {
    ' ': 'togglePlayPause',
    ArrowRight: 'seekForward',
    ArrowLeft: 'seekBackward',
    ArrowUp: 'volumeUp',
    ArrowDown: 'volumeDown',
    m: 'toggleMute',
    s: 'toggleShuffle',
    r: 'cycleRepeat',
};

const actionsByShiftedKey: Record<string, PlayerShortcutAction> = {
    ArrowRight: 'next',
    ArrowLeft: 'previous',
};

export const resolvePlayerShortcutAction = (
    event: PlayerKeyboardEventKeys
): PlayerShortcutAction | null => {
    if (event.ctrlKey || event.metaKey || event.altKey) return null;
    if (event.shiftKey) return actionsByShiftedKey[event.key] ?? null;
    return actionsByPlainKey[event.key] ?? null;
};
