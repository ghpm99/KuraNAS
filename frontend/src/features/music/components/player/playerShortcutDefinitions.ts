import type { ShortcutDefinition } from '@/components/shortcuts/shortcutDefinition';

const PLAYER_SECTION_TITLE_KEY = 'SHORTCUTS_SECTION_PLAYER';

export const playerShortcutDefinitions: ShortcutDefinition[] = [
    { keyLabels: ['Space'], descriptionKey: 'PLAYER_SHORTCUT_PLAY_PAUSE' },
    { keyLabels: ['Shift + →'], descriptionKey: 'PLAYER_SHORTCUT_NEXT' },
    { keyLabels: ['Shift + ←'], descriptionKey: 'PLAYER_SHORTCUT_PREVIOUS' },
    { keyLabels: ['→', '←'], descriptionKey: 'PLAYER_SHORTCUT_SEEK' },
    { keyLabels: ['↑', '↓'], descriptionKey: 'PLAYER_SHORTCUT_VOLUME' },
    { keyLabels: ['m'], descriptionKey: 'PLAYER_SHORTCUT_MUTE' },
    { keyLabels: ['s'], descriptionKey: 'PLAYER_SHORTCUT_SHUFFLE' },
    { keyLabels: ['r'], descriptionKey: 'PLAYER_SHORTCUT_REPEAT' },
].map((definition) => ({ ...definition, sectionTitleKey: PLAYER_SECTION_TITLE_KEY }));
