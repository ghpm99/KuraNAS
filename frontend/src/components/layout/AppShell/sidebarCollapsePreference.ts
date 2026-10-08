export const sidebarCollapsedStorageKey = 'kuranas.sidebarCollapsed';
export const sidebarWidthCssVariable = '--app-shell-sidebar-current-width';
export const compactDesktopMediaQuery = '(min-width: 900.02px) and (max-width: 1200px)';

export const readSidebarCollapsedChoice = (): boolean | null => {
    try {
        const storedChoice = window.localStorage.getItem(sidebarCollapsedStorageKey);
        if (storedChoice === 'true') return true;
        if (storedChoice === 'false') return false;
        return null;
    } catch {
        return null;
    }
};

export const writeSidebarCollapsedChoice = (isCollapsed: boolean) => {
    try {
        window.localStorage.setItem(sidebarCollapsedStorageKey, String(isCollapsed));
    } catch {
        return;
    }
};

export const isTypingTarget = (target: EventTarget | null): boolean => {
    if (!(target instanceof HTMLElement)) return false;
    if (target.isContentEditable) return true;
    return ['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName);
};
