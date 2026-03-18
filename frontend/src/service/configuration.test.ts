jest.mock('./index', () => ({
    apiBase: {
        get: jest.fn(),
        put: jest.fn(),
    },
}));

import { apiBase } from './index';
import {
    getAboutConfiguration,
    getSettingsConfiguration,
    getTranslations,
    updateSettingsConfiguration,
} from './configuration';

const mockedApi = apiBase as unknown as {
    get: jest.Mock;
    put: jest.Mock;
};

const settingsRequest = {
    library: {
        watched_paths: ['/data'],
        remember_last_location: true,
        prioritize_favorites: true,
    },
    indexing: {
        scan_on_startup: true,
        extract_metadata: true,
        generate_previews: true,
    },
    players: {
        remember_music_queue: true,
        remember_video_progress: true,
        autoplay_next_video: true,
        image_slideshow_seconds: 8,
    },
    appearance: { accent_color: 'cyan' as const, reduce_motion: true },
    language: { current: 'pt-BR' },
};

describe('service/configuration', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it.each([
        {
            name: 'gets about configuration',
            fn: () => getAboutConfiguration(),
            method: 'get' as const,
            url: '/configuration/about',
            payload: undefined,
            response: { version: '1.0.0' },
        },
        {
            name: 'gets translations',
            fn: () => getTranslations(),
            method: 'get' as const,
            url: '/configuration/translation',
            payload: undefined,
            response: { SETTINGS: 'Settings' },
        },
        {
            name: 'gets settings configuration',
            fn: () => getSettingsConfiguration(),
            method: 'get' as const,
            url: '/configuration/settings',
            payload: undefined,
            response: {
                library: {
                    runtime_root_path: '/data',
                    watched_paths: ['/data'],
                    remember_last_location: true,
                    prioritize_favorites: true,
                },
                indexing: {
                    workers_enabled: true,
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
                appearance: { accent_color: 'violet', reduce_motion: false },
                language: { current: 'en-US', available: ['en-US'] },
            },
        },
        {
            name: 'updates settings configuration',
            fn: () => updateSettingsConfiguration(settingsRequest),
            method: 'put' as const,
            url: '/configuration/settings',
            payload: settingsRequest,
            response: {
                ...settingsRequest,
                library: { ...settingsRequest.library, runtime_root_path: '/data' },
                indexing: { ...settingsRequest.indexing, workers_enabled: true },
                language: { ...settingsRequest.language, available: ['en-US', 'pt-BR'] },
            },
        },
    ])('$name', async ({ fn, method, url, payload, response }) => {
        mockedApi[method].mockResolvedValue({ data: response });

        const result = await fn();

        if (payload) {
            expect(mockedApi[method]).toHaveBeenCalledWith(url, payload);
        } else {
            expect(mockedApi[method]).toHaveBeenCalledWith(url);
        }
        expect(result).toEqual(response);
    });
});
