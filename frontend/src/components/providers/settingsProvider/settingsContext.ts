import { createContext, useContext } from 'react';
import type {
    SettingsConfiguration,
    UpdateSettingsConfigurationRequest,
} from '@/service/configuration';

export const defaultSettingsConfiguration: SettingsConfiguration = {
    library: {
        runtime_root_path: '',
        watched_paths: [],
        remember_last_location: true,
        prioritize_favorites: true,
    },
    indexing: {
        workers_enabled: false,
        scan_on_startup: true,
        extract_metadata: true,
        generate_previews: true,
    },
    players: {
        remember_music_queue: true,
        remember_video_progress: true,
        autoplay_next_video: true,
        image_slideshow_seconds: 4,
    },
    appearance: {
        accent_color: 'violet',
        reduce_motion: false,
    },
    language: {
        current: 'en-US',
        available: ['en-US'],
    },
};

export type SettingsContextType = {
    settings: SettingsConfiguration;
    isLoading: boolean;
    isSaving: boolean;
    hasError: boolean;
    refresh: () => Promise<void>;
    saveSettings: (request: UpdateSettingsConfigurationRequest) => Promise<SettingsConfiguration>;
};

const SettingsContext = createContext<SettingsContextType | undefined>(undefined);

export const SettingsContextProvider = SettingsContext.Provider;

export const useSettings = () => {
    const context = useContext(SettingsContext);
    if (!context) {
        throw new Error('useSettings must be used within a SettingsProvider');
    }
    return context;
};
