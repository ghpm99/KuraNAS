export interface PlayerShortcutHelpEntry {
    keys: string[];
    descriptionKey: string;
}

export const playerShortcutHelpEntries: PlayerShortcutHelpEntry[] = [
    { keys: ['VIDEO_SHORTCUT_KEY_SPACE', 'K'], descriptionKey: 'VIDEO_SHORTCUT_HELP_TOGGLE_PLAY' },
    { keys: ['←', '→'], descriptionKey: 'VIDEO_SHORTCUT_HELP_SEEK_SMALL' },
    { keys: ['J', 'L'], descriptionKey: 'VIDEO_SHORTCUT_HELP_SEEK_LARGE' },
    { keys: ['↑', '↓'], descriptionKey: 'VIDEO_SHORTCUT_HELP_VOLUME' },
    { keys: ['M'], descriptionKey: 'VIDEO_SHORTCUT_HELP_MUTE' },
    { keys: ['F'], descriptionKey: 'VIDEO_SHORTCUT_HELP_FULLSCREEN' },
    { keys: ['I'], descriptionKey: 'VIDEO_SHORTCUT_HELP_PICTURE_IN_PICTURE' },
    { keys: ['Shift+N'], descriptionKey: 'VIDEO_SHORTCUT_HELP_NEXT' },
    { keys: ['Shift+P'], descriptionKey: 'VIDEO_SHORTCUT_HELP_PREVIOUS' },
    { keys: ['0 – 9'], descriptionKey: 'VIDEO_SHORTCUT_HELP_SEEK_PERCENT' },
    {
        keys: ['VIDEO_SHORTCUT_KEY_HOME', 'VIDEO_SHORTCUT_KEY_END'],
        descriptionKey: 'VIDEO_SHORTCUT_HELP_START_END',
    },
    { keys: ['<', '>'], descriptionKey: 'VIDEO_SHORTCUT_HELP_SPEED' },
    { keys: ['?'], descriptionKey: 'VIDEO_SHORTCUT_HELP_OPEN_HELP' },
];
