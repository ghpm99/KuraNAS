import type { ShortcutDefinition } from './shortcutDefinition';

export const globalShortcutDefinitions: ShortcutDefinition[] = [
    { keyLabels: ['Ctrl + K', 'Cmd + K'], descriptionKey: 'SHORTCUT_GLOBAL_SEARCH' },
    { keyLabels: ['['], descriptionKey: 'SHORTCUT_TOGGLE_SIDEBAR' },
    { keyLabels: ['g h'], descriptionKey: 'SHORTCUT_GO_HOME' },
    { keyLabels: ['g f'], descriptionKey: 'SHORTCUT_GO_FILES' },
    { keyLabels: ['g i'], descriptionKey: 'SHORTCUT_GO_IMAGES' },
    { keyLabels: ['g m'], descriptionKey: 'SHORTCUT_GO_MUSIC' },
    { keyLabels: ['g v'], descriptionKey: 'SHORTCUT_GO_VIDEOS' },
    { keyLabels: ['g s'], descriptionKey: 'SHORTCUT_GO_SETTINGS' },
    { keyLabels: ['?'], descriptionKey: 'SHORTCUT_SHOW_HELP' },
];
