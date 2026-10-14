export type ShortcutTargetKind = 'editable' | 'slider' | 'button' | 'other';

export type PlayerShortcutAction =
    | { type: 'togglePlayPause' }
    | { type: 'seekRelative'; seconds: number }
    | { type: 'seekToFraction'; fraction: number }
    | { type: 'adjustVolume'; delta: number }
    | { type: 'toggleMute' }
    | { type: 'toggleFullscreen' }
    | { type: 'nextVideo' }
    | { type: 'previousVideo' }
    | { type: 'adjustPlaybackRate'; delta: number }
    | { type: 'openHelp' };

export interface ShortcutKeyInput {
    key: string;
    shiftKey: boolean;
    ctrlKey: boolean;
    metaKey: boolean;
    altKey: boolean;
    targetKind: ShortcutTargetKind;
}

export const SMALL_SEEK_SECONDS = 5;
export const LARGE_SEEK_SECONDS = 10;
export const VOLUME_STEP = 0.05;
export const PLAYBACK_RATE_STEP = 0.25;
export const MIN_PLAYBACK_RATE = 0.25;
export const MAX_PLAYBACK_RATE = 2;

const sliderOwnedKeys = new Set(['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown', 'Home', 'End']);

const symbolActions: Record<string, PlayerShortcutAction> = {
    '?': { type: 'openHelp' },
    '<': { type: 'adjustPlaybackRate', delta: -PLAYBACK_RATE_STEP },
    '>': { type: 'adjustPlaybackRate', delta: PLAYBACK_RATE_STEP },
};

const namedKeyActions: Record<string, PlayerShortcutAction> = {
    ' ': { type: 'togglePlayPause' },
    ArrowLeft: { type: 'seekRelative', seconds: -SMALL_SEEK_SECONDS },
    ArrowRight: { type: 'seekRelative', seconds: SMALL_SEEK_SECONDS },
    ArrowUp: { type: 'adjustVolume', delta: VOLUME_STEP },
    ArrowDown: { type: 'adjustVolume', delta: -VOLUME_STEP },
    Home: { type: 'seekToFraction', fraction: 0 },
    End: { type: 'seekToFraction', fraction: 1 },
};

const letterActions: Record<string, PlayerShortcutAction> = {
    k: { type: 'togglePlayPause' },
    j: { type: 'seekRelative', seconds: -LARGE_SEEK_SECONDS },
    l: { type: 'seekRelative', seconds: LARGE_SEEK_SECONDS },
    m: { type: 'toggleMute' },
    f: { type: 'toggleFullscreen' },
};

const shiftedLetterActions: Record<string, PlayerShortcutAction> = {
    n: { type: 'nextVideo' },
    p: { type: 'previousVideo' },
};

const isDigitKey = (key: string) => key.length === 1 && key >= '0' && key <= '9';

const isEditableElement = (element: HTMLElement) => {
    if (['INPUT', 'TEXTAREA', 'SELECT'].includes(element.tagName)) {
        return true;
    }
    const contentEditableAttribute = element.getAttribute('contenteditable');
    return (
        element.isContentEditable === true ||
        contentEditableAttribute === '' ||
        contentEditableAttribute === 'true'
    );
};

export function classifyShortcutTarget(target: EventTarget | null): ShortcutTargetKind {
    if (!(target instanceof HTMLElement)) {
        return 'other';
    }
    const role = target.getAttribute('role');
    if (
        role === 'slider' ||
        (target.tagName === 'INPUT' && target.getAttribute('type') === 'range')
    ) {
        return 'slider';
    }
    if (isEditableElement(target)) {
        return 'editable';
    }
    if (target.tagName === 'BUTTON' || target.tagName === 'A' || role === 'button') {
        return 'button';
    }
    return 'other';
}

export function resolvePlayerShortcut(input: ShortcutKeyInput): PlayerShortcutAction | null {
    const { key, shiftKey, targetKind } = input;
    if (input.ctrlKey || input.metaKey || input.altKey) return null;
    if (targetKind === 'editable') return null;
    if (targetKind === 'slider' && sliderOwnedKeys.has(key)) return null;
    if (targetKind === 'button' && key === ' ') return null;

    const symbolAction = symbolActions[key];
    if (symbolAction) return symbolAction;

    const lowerKey = key.toLowerCase();
    if (shiftKey) return shiftedLetterActions[lowerKey] ?? null;

    if (isDigitKey(key)) return { type: 'seekToFraction', fraction: Number(key) / 10 };
    return namedKeyActions[key] ?? letterActions[lowerKey] ?? null;
}
