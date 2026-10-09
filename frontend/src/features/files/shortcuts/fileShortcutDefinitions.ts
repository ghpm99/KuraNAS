import type { ShortcutDefinition } from '@/components/shortcuts/shortcutDefinition';

export const fileShortcutDefinitions: ShortcutDefinition[] = [
    { keyLabels: ['Enter'], descriptionKey: 'FILES_SHORTCUT_OPEN' },
    { keyLabels: ['Delete'], descriptionKey: 'FILES_SHORTCUT_DELETE' },
    { keyLabels: ['F2'], descriptionKey: 'FILES_SHORTCUT_RENAME' },
    { keyLabels: ['Ctrl + A', 'Cmd + A'], descriptionKey: 'FILES_SHORTCUT_SELECT_ALL' },
    { keyLabels: ['Esc'], descriptionKey: 'FILES_SHORTCUT_CLEAR_SELECTION' },
    { keyLabels: ['Backspace', 'Alt + ↑'], descriptionKey: 'FILES_SHORTCUT_PARENT_FOLDER' },
    { keyLabels: ['/'], descriptionKey: 'FILES_SHORTCUT_FOCUS_SEARCH' },
    { keyLabels: ['← ↑ → ↓'], descriptionKey: 'FILES_SHORTCUT_MOVE_FOCUS' },
    { keyLabels: ['Shift + ← ↑ → ↓'], descriptionKey: 'FILES_SHORTCUT_EXTEND_SELECTION' },
];
