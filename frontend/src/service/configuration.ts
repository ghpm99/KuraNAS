import { AboutContextType } from '@/components/providers/aboutProvider/AboutContext';
import { apiBase } from '.';

export type SettingsConfiguration = {
    indexing: {
        workers_enabled: boolean;
        scan_on_startup: boolean;
        extract_metadata: boolean;
        generate_previews: boolean;
    };
    captures: {
        save_path: string;
        default_path: string;
        storage_roots: string[];
    };
    ai: {
        image_classification: boolean;
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
    indexing: {
        scan_on_startup: boolean;
        extract_metadata: boolean;
        generate_previews: boolean;
    };
    captures: {
        save_path: string;
    };
    ai: {
        image_classification: boolean;
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

export type EnvFieldKind = 'string' | 'int' | 'bool' | 'secret';

export type EnvField = {
    key: string;
    group: string;
    kind: EnvFieldKind;
    value: string;
    configured: boolean;
    dangerous: boolean;
};

export type EnvConfig = {
    fields: EnvField[];
    restart_required: boolean;
};

export type UpdateEnvConfigRequest = {
    changes: Record<string, string>;
    confirmed: boolean;
};

export type EnvTestResult = {
    ok: boolean;
    message: string;
};

export type TestDatabaseRequest = {
    host: string;
    port: string;
    user: string;
    name: string;
    password: string;
};

export type TestPathRequest = {
    path: string;
};

export const getEnvConfig = async (): Promise<EnvConfig> => {
    const response = await apiBase.get<EnvConfig>('/configuration/env');
    return response.data;
};

export const updateEnvConfig = async (request: UpdateEnvConfigRequest): Promise<EnvConfig> => {
    const response = await apiBase.put<EnvConfig>('/configuration/env', request);
    return response.data;
};

export const testEnvDatabase = async (request: TestDatabaseRequest): Promise<EnvTestResult> => {
    const response = await apiBase.post<EnvTestResult>('/configuration/env/test-db', request);
    return response.data;
};

export const testEnvPath = async (request: TestPathRequest): Promise<EnvTestResult> => {
    const response = await apiBase.post<EnvTestResult>('/configuration/env/test-path', request);
    return response.data;
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
