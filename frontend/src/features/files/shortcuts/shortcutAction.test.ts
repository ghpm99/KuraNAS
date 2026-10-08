import { resolveNextFocusIndex, resolveShortcutAction } from './shortcutAction';

const keys = (key: string, modifiers: Partial<KeyboardEvent> = {}) => ({
    key,
    ctrlKey: false,
    metaKey: false,
    altKey: false,
    shiftKey: false,
    ...modifiers,
});

describe('resolveShortcutAction', () => {
    it.each([
        ['Enter', { kind: 'openItem' }],
        ['Delete', { kind: 'deleteSelection' }],
        ['F2', { kind: 'renameSelection' }],
        ['Backspace', { kind: 'goToParent' }],
        ['/', { kind: 'focusSearch' }],
    ])('maps %s to its action', (key, expectedAction) => {
        expect(resolveShortcutAction(keys(key))).toEqual(expectedAction);
    });

    it('maps ? (which needs shift) to the help action', () => {
        expect(resolveShortcutAction(keys('?', { shiftKey: true }))).toEqual({ kind: 'showHelp' });
    });

    it('maps Ctrl+A and Cmd+A to select all', () => {
        expect(resolveShortcutAction(keys('a', { ctrlKey: true }))).toEqual({ kind: 'selectAll' });
        expect(resolveShortcutAction(keys('A', { metaKey: true }))).toEqual({ kind: 'selectAll' });
    });

    it('maps Alt+ArrowUp to parent folder and ignores other alt combinations', () => {
        expect(resolveShortcutAction(keys('ArrowUp', { altKey: true }))).toEqual({
            kind: 'goToParent',
        });
        expect(resolveShortcutAction(keys('ArrowDown', { altKey: true }))).toBeNull();
    });

    it('maps arrows to focus moves and shift+arrows to selection extension', () => {
        expect(resolveShortcutAction(keys('ArrowLeft'))).toEqual({
            kind: 'moveFocus',
            direction: 'left',
            isExtendingSelection: false,
        });
        expect(resolveShortcutAction(keys('ArrowDown', { shiftKey: true }))).toEqual({
            kind: 'moveFocus',
            direction: 'down',
            isExtendingSelection: true,
        });
    });

    it('ignores unmapped keys, other ctrl combinations and shifted plain keys', () => {
        expect(resolveShortcutAction(keys('x'))).toBeNull();
        expect(resolveShortcutAction(keys('c', { ctrlKey: true }))).toBeNull();
        expect(resolveShortcutAction(keys('Delete', { shiftKey: true }))).toBeNull();
    });
});

describe('resolveNextFocusIndex', () => {
    const grid = { columnCount: 3, itemCount: 8 };

    it('starts at the first item when nothing is focused', () => {
        expect(resolveNextFocusIndex({ currentIndex: -1, direction: 'down', ...grid })).toBe(0);
    });

    it('moves by one horizontally and by the column count vertically', () => {
        expect(resolveNextFocusIndex({ currentIndex: 4, direction: 'right', ...grid })).toBe(5);
        expect(resolveNextFocusIndex({ currentIndex: 4, direction: 'left', ...grid })).toBe(3);
        expect(resolveNextFocusIndex({ currentIndex: 4, direction: 'down', ...grid })).toBe(7);
        expect(resolveNextFocusIndex({ currentIndex: 4, direction: 'up', ...grid })).toBe(1);
    });

    it('clamps at both ends', () => {
        expect(resolveNextFocusIndex({ currentIndex: 0, direction: 'up', ...grid })).toBe(0);
        expect(resolveNextFocusIndex({ currentIndex: 7, direction: 'down', ...grid })).toBe(7);
    });

    it('returns -1 for an empty list', () => {
        expect(
            resolveNextFocusIndex({ currentIndex: -1, direction: 'down', columnCount: 1, itemCount: 0 })
        ).toBe(-1);
    });
});
