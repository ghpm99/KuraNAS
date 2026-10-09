import { resolvePlayerShortcutAction } from './playerShortcutAction';

const keyEvent = (key: string, modifiers: Partial<KeyboardEvent> = {}) => ({
    key,
    ctrlKey: false,
    metaKey: false,
    altKey: false,
    shiftKey: false,
    ...modifiers,
});

describe('resolvePlayerShortcutAction', () => {
    it.each([
        [' ', 'togglePlayPause'],
        ['ArrowRight', 'seekForward'],
        ['ArrowLeft', 'seekBackward'],
        ['ArrowUp', 'volumeUp'],
        ['ArrowDown', 'volumeDown'],
        ['m', 'toggleMute'],
        ['s', 'toggleShuffle'],
        ['r', 'cycleRepeat'],
    ])('maps plain %j to %s', (key, expectedAction) => {
        expect(resolvePlayerShortcutAction(keyEvent(key))).toBe(expectedAction);
    });

    it('maps shifted horizontal arrows to next and previous', () => {
        expect(resolvePlayerShortcutAction(keyEvent('ArrowRight', { shiftKey: true }))).toBe(
            'next'
        );
        expect(resolvePlayerShortcutAction(keyEvent('ArrowLeft', { shiftKey: true }))).toBe(
            'previous'
        );
    });

    it('ignores shifted keys that have no shifted action', () => {
        expect(resolvePlayerShortcutAction(keyEvent('ArrowUp', { shiftKey: true }))).toBeNull();
        expect(resolvePlayerShortcutAction(keyEvent('m', { shiftKey: true }))).toBeNull();
    });

    it.each(['ctrlKey', 'metaKey', 'altKey'] as const)('ignores %s combinations', (modifier) => {
        expect(resolvePlayerShortcutAction(keyEvent(' ', { [modifier]: true }))).toBeNull();
    });

    it('ignores unrelated keys', () => {
        expect(resolvePlayerShortcutAction(keyEvent('x'))).toBeNull();
    });
});
