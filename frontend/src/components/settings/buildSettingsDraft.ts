import type {
    SettingsConfiguration,
    UpdateSettingsConfigurationRequest,
} from '@/service/configuration';

export const buildDraftFromSettings = (
    settings: SettingsConfiguration
): UpdateSettingsConfigurationRequest => ({
    indexing: {
        scan_on_startup: settings.indexing.scan_on_startup,
        extract_metadata: settings.indexing.extract_metadata,
        generate_previews: settings.indexing.generate_previews,
    },
    captures: {
        save_path: settings.captures.save_path,
    },
    ai: {
        image_classification: settings.ai.image_classification,
    },
    players: {
        remember_music_queue: settings.players.remember_music_queue,
        remember_video_progress: settings.players.remember_video_progress,
        autoplay_next_video: settings.players.autoplay_next_video,
        image_slideshow_seconds: settings.players.image_slideshow_seconds,
    },
    appearance: {
        accent_color: settings.appearance.accent_color,
        reduce_motion: settings.appearance.reduce_motion,
        theme_mode: settings.appearance.theme_mode,
    },
    language: {
        current: settings.language.current,
    },
});
