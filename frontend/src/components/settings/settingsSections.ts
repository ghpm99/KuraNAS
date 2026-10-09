import { appRoutes } from '@/app/routes';

export type SettingsSectionDefinition = {
    id: string;
    labelKey: string;
    descriptionKey: string;
};

export const settingsSections = [
    {
        id: 'storage-roots',
        labelKey: 'SETTINGS_STORAGE_ROOTS_TITLE',
        descriptionKey: 'SETTINGS_STORAGE_ROOTS_DESCRIPTION',
    },
    {
        id: 'backup',
        labelKey: 'SETTINGS_BACKUP_TITLE',
        descriptionKey: 'SETTINGS_BACKUP_DESCRIPTION',
    },
    {
        id: 'tiering',
        labelKey: 'SETTINGS_TIERING_TITLE',
        descriptionKey: 'SETTINGS_TIERING_DESCRIPTION',
    },
    {
        id: 'auto-shutdown',
        labelKey: 'SETTINGS_AUTO_SHUTDOWN_TITLE',
        descriptionKey: 'SETTINGS_AUTO_SHUTDOWN_DESCRIPTION',
    },
    {
        id: 'access-control',
        labelKey: 'SETTINGS_ACCESS_CONTROL_TITLE',
        descriptionKey: 'SETTINGS_ACCESS_CONTROL_DESCRIPTION',
    },
    {
        id: 'libraries',
        labelKey: 'SETTINGS_LIBRARIES_TITLE',
        descriptionKey: 'SETTINGS_LIBRARIES_DESCRIPTION',
    },
    {
        id: 'ai-providers',
        labelKey: 'AI_PROVIDERS_TITLE',
        descriptionKey: 'AI_PROVIDERS_DESCRIPTION',
    },
    {
        id: 'yt-dlp',
        labelKey: 'SETTINGS_YTDLP_TITLE',
        descriptionKey: 'SETTINGS_YTDLP_DESCRIPTION',
    },
    {
        id: 'email',
        labelKey: 'SETTINGS_EMAIL_TITLE',
        descriptionKey: 'SETTINGS_EMAIL_HELP',
    },
    {
        id: 'indexing',
        labelKey: 'SETTINGS_SECTION_INDEXING',
        descriptionKey: 'SETTINGS_SECTION_INDEXING_DESCRIPTION',
    },
    {
        id: 'captures',
        labelKey: 'SETTINGS_SECTION_CAPTURES',
        descriptionKey: 'SETTINGS_SECTION_CAPTURES_DESCRIPTION',
    },
    {
        id: 'ai',
        labelKey: 'SETTINGS_SECTION_AI',
        descriptionKey: 'SETTINGS_SECTION_AI_DESCRIPTION',
    },
    {
        id: 'players',
        labelKey: 'SETTINGS_SECTION_PLAYERS',
        descriptionKey: 'SETTINGS_SECTION_PLAYERS_DESCRIPTION',
    },
    {
        id: 'appearance',
        labelKey: 'SETTINGS_SECTION_APPEARANCE',
        descriptionKey: 'SETTINGS_SECTION_APPEARANCE_DESCRIPTION',
    },
    {
        id: 'language',
        labelKey: 'SETTINGS_SECTION_LANGUAGE',
        descriptionKey: 'SETTINGS_SECTION_LANGUAGE_DESCRIPTION',
    },
] as const satisfies readonly SettingsSectionDefinition[];

export type SettingsSectionId = (typeof settingsSections)[number]['id'];

export const getSettingsSectionRoute = (sectionId: SettingsSectionId) =>
    `${appRoutes.settings}#${sectionId}`;
