import { AboutContextType } from '@/components/providers/aboutProvider/AboutContext';
import { apiBase } from '.';

export type SettingsConfiguration = {
    library: {
        runtime_root_path: string;
        watched_paths: string[];
        remember_last_location: boolean;
        prioritize_favorites: boolean;
    };
    indexing: {
        workers_enabled: boolean;
        scan_on_startup: boolean;
        extract_metadata: boolean;
        generate_previews: boolean;
    };
    players: {
        remember_music_queue: boolean;
        remember_video_progress: boolean;
        autoplay_next_video: boolean;
        image_slideshow_seconds: number;
    };
    appearance: {
        accent_color: 'violet' | 'cyan' | 'rose';
        reduce_motion: boolean;
    };
    language: {
        current: string;
        available: string[];
    };
};

export type UpdateSettingsConfigurationRequest = {
    library: {
        watched_paths: string[];
        remember_last_location: boolean;
        prioritize_favorites: boolean;
    };
    indexing: {
        scan_on_startup: boolean;
        extract_metadata: boolean;
        generate_previews: boolean;
    };
    players: {
        remember_music_queue: boolean;
        remember_video_progress: boolean;
        autoplay_next_video: boolean;
        image_slideshow_seconds: number;
    };
    appearance: {
        accent_color: 'violet' | 'cyan' | 'rose';
        reduce_motion: boolean;
    };
    language: {
        current: string;
    };
};

export const getAboutConfiguration = async (): Promise<AboutContextType> => {
    const response = await apiBase.get<AboutContextType>('/configuration/about');
    return response.data;
};

export const getTranslations = async (): Promise<Record<string, string>> => {
    const response = await apiBase.get<Record<string, string>>('/configuration/translation');
    return response.data;
};

export const getSettingsConfiguration = async (): Promise<SettingsConfiguration> => {
    const response = await apiBase.get<SettingsConfiguration>('/configuration/settings');
    return response.data;
};

export const updateSettingsConfiguration = async (
    request: UpdateSettingsConfigurationRequest
): Promise<SettingsConfiguration> => {
    const response = await apiBase.put<SettingsConfiguration>('/configuration/settings', request);
    return response.data;
};
