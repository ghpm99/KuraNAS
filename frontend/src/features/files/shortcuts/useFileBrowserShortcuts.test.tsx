import { act, fireEvent, renderHook } from '@testing-library/react';
import { useFileSelection } from '@/features/files/selection/useFileSelection';
import { createTestFile } from '@/features/files/selection/testFileFactory';
import useFileBrowserShortcuts from './useFileBrowserShortcuts';

const listedFiles = [1, 2, 3, 4, 5, 6].map((id) => createTestFile(id));

const setUpHook = (overrides: Partial<Parameters<typeof useFileBrowserShortcuts>[0]> = {}) => {
    const handlers = {
        onOpenFile: jest.fn(),
        onDeleteFiles: jest.fn(),
        onRenameFile: jest.fn(),
        onGoToParent: jest.fn(),
        onFocusSearch: jest.fn(),
    };
    const rendered = renderHook(() => {
        const selection = useFileSelection('scope');
        const shortcuts = useFileBrowserShortcuts({
            isEnabled: true,
            files: listedFiles,
            selection,
            getColumnCount: () => 3,
            ...handlers,
            ...overrides,
        });
        return { selection, shortcuts };
    });
    return { ...rendered, handlers };
};

const press = (key: string, init: KeyboardEventInit = {}, target: Element | Document = document) => {
    act(() => {
        fireEvent.keyDown(target, { key, ...init });
    });
};

describe('useFileBrowserShortcuts', () => {
    afterEach(() => {
        document.body.innerHTML = '';
    });

    it('mounts with an empty list and no handlers without crashing', () => {
        const { result } = renderHook(() => {
            const selection = useFileSelection('scope');
            return useFileBrowserShortcuts({
                isEnabled: true,
                files: [],
                selection,
                getColumnCount: () => 1,
                onOpenFile: jest.fn(),
                onDeleteFiles: jest.fn(),
                onRenameFile: jest.fn(),
            });
        });

        press('ArrowDown');
        press('Enter');
        press('Delete');
        press('Backspace');
        press('/');

        expect(result.current.tabStopFileId).toBeNull();
        expect(result.current.isHelpOpen).toBe(false);
    });

    it('makes the first item the tab stop until focus moves', () => {
        const { result } = setUpHook();

        expect(result.current.shortcuts.tabStopFileId).toBe(1);
    });

    it('moves the focus ring with arrows by one horizontally and by columns vertically', () => {
        const { result } = setUpHook();

        press('ArrowDown');
        expect(result.current.shortcuts.tabStopFileId).toBe(1);
        press('ArrowRight');
        expect(result.current.shortcuts.tabStopFileId).toBe(2);
        press('ArrowDown');
        expect(result.current.shortcuts.tabStopFileId).toBe(5);
        press('ArrowLeft');
        expect(result.current.shortcuts.tabStopFileId).toBe(4);
        press('ArrowUp');
        expect(result.current.shortcuts.tabStopFileId).toBe(1);
    });

    it('moves the DOM focus to the rendered item', () => {
        const item = document.createElement('a');
        item.setAttribute('data-file-id', '2');
        item.href = '/f/2';
        document.body.appendChild(item);
        setUpHook();

        press('ArrowDown');
        press('ArrowRight');

        expect(document.activeElement).toBe(item);
    });

    it('extends the selection with shift+arrows starting from the focused item', () => {
        const { result } = setUpHook();

        press('ArrowDown');
        press('ArrowRight', { shiftKey: true });
        press('ArrowRight', { shiftKey: true });

        expect(result.current.selection.selectedFiles.map((file) => file.id).sort()).toEqual([1, 2, 3]);
    });

    it('opens the focused item with Enter and falls back to the single selected item', () => {
        const { result, handlers } = setUpHook();

        act(() => result.current.selection.toggle(listedFiles[3]!));
        press('Enter');
        expect(handlers.onOpenFile).toHaveBeenLastCalledWith(listedFiles[3]);

        press('ArrowDown');
        press('Enter');
        expect(handlers.onOpenFile).toHaveBeenLastCalledWith(listedFiles[0]);
    });

    it('does not open anything on Enter when nothing is focused or selected', () => {
        const { handlers } = setUpHook();

        press('Enter');

        expect(handlers.onOpenFile).not.toHaveBeenCalled();
    });

    it('lets a focused link handle Enter natively', () => {
        const link = document.createElement('a');
        link.href = '/f/1';
        document.body.appendChild(link);
        const { handlers } = setUpHook();
        press('ArrowDown');

        press('Enter', {}, link);

        expect(handlers.onOpenFile).not.toHaveBeenCalled();
    });

    it('deletes the whole selection, or the focused item when nothing is selected', () => {
        const { result, handlers } = setUpHook();

        press('ArrowDown');
        press('Delete');
        expect(handlers.onDeleteFiles).toHaveBeenLastCalledWith([listedFiles[0]]);

        act(() => result.current.selection.selectAll(listedFiles.slice(0, 2)));
        press('Delete');
        expect(handlers.onDeleteFiles).toHaveBeenLastCalledWith(listedFiles.slice(0, 2));
    });

    it('does not delete with nothing focused or selected', () => {
        const { handlers } = setUpHook();

        press('Delete');

        expect(handlers.onDeleteFiles).not.toHaveBeenCalled();
    });

    it('renames only a single selection or the focused item', () => {
        const { result, handlers } = setUpHook();

        act(() => result.current.selection.selectAll(listedFiles.slice(0, 2)));
        press('F2');
        expect(handlers.onRenameFile).not.toHaveBeenCalled();

        act(() => result.current.selection.clear());
        act(() => result.current.selection.toggle(listedFiles[2]!));
        press('F2');
        expect(handlers.onRenameFile).toHaveBeenLastCalledWith(listedFiles[2]);

        act(() => result.current.selection.clear());
        press('ArrowDown');
        press('F2');
        expect(handlers.onRenameFile).toHaveBeenLastCalledWith(listedFiles[0]);
    });

    it('selects all loaded items with Ctrl+A and Cmd+A', () => {
        const { result } = setUpHook();

        press('a', { ctrlKey: true });
        expect(result.current.selection.selectedCount).toBe(6);

        act(() => result.current.selection.clear());
        press('a', { metaKey: true });
        expect(result.current.selection.selectedCount).toBe(6);
    });

    it('goes to the parent folder with Backspace and Alt+ArrowUp', () => {
        const { handlers } = setUpHook();

        press('Backspace');
        press('ArrowUp', { altKey: true });

        expect(handlers.onGoToParent).toHaveBeenCalledTimes(2);
    });

    it('focuses the search with / and opens then closes the help with ?', () => {
        const { result, handlers } = setUpHook();

        press('/');
        expect(handlers.onFocusSearch).toHaveBeenCalledTimes(1);

        press('?', { shiftKey: true });
        expect(result.current.shortcuts.isHelpOpen).toBe(true);
        act(() => result.current.shortcuts.closeHelp());
        expect(result.current.shortcuts.isHelpOpen).toBe(false);
    });

    it('ignores every shortcut while typing in an input', () => {
        const input = document.createElement('input');
        document.body.appendChild(input);
        const { result, handlers } = setUpHook();

        press('Backspace', {}, input);
        press('/', {}, input);
        press('a', { ctrlKey: true }, input);
        press('Delete', {}, input);

        expect(handlers.onGoToParent).not.toHaveBeenCalled();
        expect(handlers.onFocusSearch).not.toHaveBeenCalled();
        expect(handlers.onDeleteFiles).not.toHaveBeenCalled();
        expect(result.current.selection.selectedCount).toBe(0);
    });

    it('ignores every shortcut while a dialog is open', () => {
        const modal = document.createElement('div');
        modal.className = 'MuiModal-root';
        document.body.appendChild(modal);
        const { handlers } = setUpHook();

        press('Backspace');
        press('/');

        expect(handlers.onGoToParent).not.toHaveBeenCalled();
        expect(handlers.onFocusSearch).not.toHaveBeenCalled();
    });

    it('does nothing when disabled', () => {
        const { handlers } = setUpHook({ isEnabled: false });

        press('Backspace');

        expect(handlers.onGoToParent).not.toHaveBeenCalled();
    });

    it('ignores Backspace and / when the screen provides no handler for them', () => {
        const { result } = setUpHook({ onGoToParent: undefined, onFocusSearch: undefined });

        press('Backspace');
        press('/');

        expect(result.current.shortcuts.isHelpOpen).toBe(false);
    });
});
