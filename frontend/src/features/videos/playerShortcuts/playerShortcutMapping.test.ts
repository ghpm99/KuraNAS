import {
    classifyShortcutTarget,
    resolvePlayerShortcut,
    type PlayerShortcutAction,
    type ShortcutKeyInput,
} from './playerShortcutMapping';

const press = (key: string, overrides: Partial<ShortcutKeyInput> = {}) =>
    resolvePlayerShortcut({
        key,
        shiftKey: false,
        ctrlKey: false,
        metaKey: false,
        altKey: false,
        targetKind: 'other',
        ...overrides,
    });

describe('features/videos/playerShortcuts/playerShortcutMapping', () => {
    it.each<[string, PlayerShortcutAction]>([
        [' ', { type: 'togglePlayPause' }],
        ['k', { type: 'togglePlayPause' }],
        ['K', { type: 'togglePlayPause' }],
        ['ArrowLeft', { type: 'seekRelative', seconds: -5 }],
        ['ArrowRight', { type: 'seekRelative', seconds: 5 }],
        ['j', { type: 'seekRelative', seconds: -10 }],
        ['l', { type: 'seekRelative', seconds: 10 }],
        ['ArrowUp', { type: 'adjustVolume', delta: 0.05 }],
        ['ArrowDown', { type: 'adjustVolume', delta: -0.05 }],
        ['m', { type: 'toggleMute' }],
        ['f', { type: 'toggleFullscreen' }],
        ['i', { type: 'togglePictureInPicture' }],
        ['I', { type: 'togglePictureInPicture' }],
        ['c', { type: 'toggleSubtitles' }],
        ['C', { type: 'toggleSubtitles' }],
        ['0', { type: 'seekToFraction', fraction: 0 }],
        ['5', { type: 'seekToFraction', fraction: 0.5 }],
        ['9', { type: 'seekToFraction', fraction: 0.9 }],
        ['Home', { type: 'seekToFraction', fraction: 0 }],
        ['End', { type: 'seekToFraction', fraction: 1 }],
        ['<', { type: 'adjustPlaybackRate', delta: -0.25 }],
        ['>', { type: 'adjustPlaybackRate', delta: 0.25 }],
        ['?', { type: 'openHelp' }],
    ])('maps %j to its action', (key, expectedAction) => {
        expect(press(key)).toEqual(expectedAction);
    });

    it('maps Shift+N and Shift+P to next and previous', () => {
        expect(press('N', { shiftKey: true })).toEqual({ type: 'nextVideo' });
        expect(press('P', { shiftKey: true })).toEqual({ type: 'previousVideo' });
    });

    it('does not map N and P without Shift', () => {
        expect(press('n')).toBeNull();
        expect(press('p')).toBeNull();
    });

    it('maps symbol shortcuts regardless of Shift state', () => {
        expect(press('?', { shiftKey: true })).toEqual({ type: 'openHelp' });
        expect(press('>', { shiftKey: true })).toEqual({
            type: 'adjustPlaybackRate',
            delta: 0.25,
        });
    });

    it('ignores unmapped keys and shifted digits and letters', () => {
        expect(press('x')).toBeNull();
        expect(press('Enter')).toBeNull();
        expect(press('!', { shiftKey: true })).toBeNull();
        expect(press('k', { shiftKey: true })).toBeNull();
    });

    it.each(['ctrlKey', 'metaKey', 'altKey'] as const)('ignores shortcuts with %s', (modifier) => {
        expect(press('k', { [modifier]: true })).toBeNull();
        expect(press('?', { [modifier]: true })).toBeNull();
        expect(press(' ', { [modifier]: true })).toBeNull();
    });

    it('ignores every shortcut while an editable field is focused', () => {
        expect(press(' ', { targetKind: 'editable' })).toBeNull();
        expect(press('k', { targetKind: 'editable' })).toBeNull();
        expect(press('?', { targetKind: 'editable' })).toBeNull();
    });

    it('leaves slider navigation keys to the focused slider', () => {
        ['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown', 'Home', 'End'].forEach((key) => {
            expect(press(key, { targetKind: 'slider' })).toBeNull();
        });
        expect(press('k', { targetKind: 'slider' })).toEqual({ type: 'togglePlayPause' });
    });

    it('leaves Space to a focused button', () => {
        expect(press(' ', { targetKind: 'button' })).toBeNull();
        expect(press('k', { targetKind: 'button' })).toEqual({ type: 'togglePlayPause' });
        expect(press('ArrowLeft', { targetKind: 'button' })).toEqual({
            type: 'seekRelative',
            seconds: -5,
        });
    });
});

describe('classifyShortcutTarget', () => {
    const build = (html: string) => {
        const container = document.createElement('div');
        container.innerHTML = html;
        return container.firstElementChild as HTMLElement;
    };

    it('classifies each kind of focus target', () => {
        expect(classifyShortcutTarget(null)).toBe('other');
        expect(classifyShortcutTarget(document)).toBe('other');
        expect(classifyShortcutTarget(build('<input type="text" />'))).toBe('editable');
        expect(classifyShortcutTarget(build('<textarea></textarea>'))).toBe('editable');
        expect(classifyShortcutTarget(build('<select></select>'))).toBe('editable');
        expect(classifyShortcutTarget(build('<div contenteditable="true"></div>'))).toBe(
            'editable'
        );
        expect(classifyShortcutTarget(build('<div contenteditable=""></div>'))).toBe('editable');
        expect(classifyShortcutTarget(build('<input type="range" />'))).toBe('slider');
        expect(classifyShortcutTarget(build('<span role="slider"></span>'))).toBe('slider');
        expect(classifyShortcutTarget(build('<button></button>'))).toBe('button');
        expect(classifyShortcutTarget(build('<a href="#x">x</a>'))).toBe('button');
        expect(classifyShortcutTarget(build('<div role="button"></div>'))).toBe('button');
        expect(classifyShortcutTarget(build('<div></div>'))).toBe('other');
    });
});
