import { act, fireEvent, render, renderHook, screen } from '@testing-library/react';
import { readFileSync } from 'fs';
import { join } from 'path';
import { useSidebarCollapse } from './useSidebarCollapse';
import { sidebarCollapsedStorageKey, sidebarWidthCssVariable } from './sidebarCollapsePreference';

const installCompactDesktopViewport = (isCompactDesktop: boolean) => {
    Object.defineProperty(window, 'matchMedia', {
        configurable: true,
        writable: true,
        value: (query: string) => ({
            matches: isCompactDesktop,
            media: query,
            addEventListener: jest.fn(),
            removeEventListener: jest.fn(),
            addListener: jest.fn(),
            removeListener: jest.fn(),
        }),
    });
};

const readSidebarWidthVariable = () =>
    document.documentElement.style.getPropertyValue(sidebarWidthCssVariable);

describe('layout/AppShell/useSidebarCollapse', () => {
    const originalMatchMedia = window.matchMedia;

    beforeEach(() => {
        window.localStorage.clear();
        document.documentElement.removeAttribute('style');
    });

    afterEach(() => {
        Object.defineProperty(window, 'matchMedia', {
            configurable: true,
            writable: true,
            value: originalMatchMedia,
        });
        jest.restoreAllMocks();
    });

    it('starts expanded without storage and without matchMedia', () => {
        const { result } = renderHook(() => useSidebarCollapse());

        expect(result.current.isCollapsed).toBe(false);
        expect(readSidebarWidthVariable()).toBe('var(--app-shell-sidebar-width)');
    });

    it('defaults to collapsed between 900px and 1200px when the user has not chosen', () => {
        installCompactDesktopViewport(true);

        const { result } = renderHook(() => useSidebarCollapse());

        expect(result.current.isCollapsed).toBe(true);
        expect(readSidebarWidthVariable()).toBe('var(--app-shell-sidebar-width-collapsed)');
    });

    it('keeps the user choice over the viewport default', () => {
        installCompactDesktopViewport(true);
        window.localStorage.setItem(sidebarCollapsedStorageKey, 'false');

        const { result } = renderHook(() => useSidebarCollapse());

        expect(result.current.isCollapsed).toBe(false);
    });

    it('persists the toggled choice per device and restores it on the next mount', () => {
        const firstMount = renderHook(() => useSidebarCollapse());

        act(() => firstMount.result.current.toggleCollapsed());

        expect(firstMount.result.current.isCollapsed).toBe(true);
        expect(window.localStorage.getItem(sidebarCollapsedStorageKey)).toBe('true');
        expect(readSidebarWidthVariable()).toBe('var(--app-shell-sidebar-width-collapsed)');
        firstMount.unmount();
        expect(readSidebarWidthVariable()).toBe('');

        const secondMount = renderHook(() => useSidebarCollapse());
        expect(secondMount.result.current.isCollapsed).toBe(true);
    });

    it('survives a storage that throws', () => {
        jest.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('blocked');
        });
        jest.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('blocked');
        });

        const { result } = renderHook(() => useSidebarCollapse());
        act(() => result.current.toggleCollapsed());

        expect(result.current.isCollapsed).toBe(true);
    });

    it('toggles with the [ key', () => {
        const { result } = renderHook(() => useSidebarCollapse());

        act(() => {
            fireEvent.keyDown(document.body, { key: '[' });
        });

        expect(result.current.isCollapsed).toBe(true);
    });

    it('ignores [ while typing, with modifiers, or for other keys', () => {
        const { result } = renderHook(() => useSidebarCollapse());
        render(
            <div>
                <input aria-label="campo" />
                <textarea aria-label="area" />
                <div contentEditable suppressContentEditableWarning aria-label="editavel" />
            </div>
        );

        fireEvent.keyDown(screen.getByLabelText('campo'), { key: '[' });
        fireEvent.keyDown(screen.getByLabelText('area'), { key: '[' });
        const editable = screen.getByLabelText('editavel');
        Object.defineProperty(editable, 'isContentEditable', { value: true });
        fireEvent.keyDown(editable, { key: '[' });
        fireEvent.keyDown(document.body, { key: '[', ctrlKey: true });
        fireEvent.keyDown(document.body, { key: ']' });

        expect(result.current.isCollapsed).toBe(false);
    });
});

describe('sidebar width variable consumers', () => {
    const readSource = (relativePath: string) =>
        readFileSync(join(__dirname, relativePath), 'utf-8');

    it('moves the shell grid column with the current sidebar width', () => {
        expect(readSource('./AppShell.module.css')).toContain(sidebarWidthCssVariable);
    });

    it('offsets the mini-player left edge by the current sidebar width', () => {
        const playerStyles = readSource(
            '../../../features/music/components/player/GlobalPlayerControl.module.css'
        );

        expect(playerStyles).toContain(`left: var(${sidebarWidthCssVariable}`);
    });
});
