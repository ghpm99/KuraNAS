const toggleSidebarEventName = 'kuranas:toggle-sidebar';
const showShortcutsHelpEventName = 'kuranas:show-shortcuts-help';

const subscribeToCommand = (eventName: string, onCommand: () => void) => {
    window.addEventListener(eventName, onCommand);
    return () => window.removeEventListener(eventName, onCommand);
};

export const requestSidebarToggle = () => window.dispatchEvent(new Event(toggleSidebarEventName));

export const requestShortcutsHelp = () =>
    window.dispatchEvent(new Event(showShortcutsHelpEventName));

export const subscribeToSidebarToggle = (onToggle: () => void) =>
    subscribeToCommand(toggleSidebarEventName, onToggle);

export const subscribeToShortcutsHelp = (onShowHelp: () => void) =>
    subscribeToCommand(showShortcutsHelpEventName, onShowHelp);
