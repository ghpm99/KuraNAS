jest.mock('./index', () => ({
    apiBase: {
        get: jest.fn(),
        put: jest.fn(),
        post: jest.fn(),
    },
}));

import { apiBase } from './index';
import {
    getAboutConfiguration,
    getEnvConfig,
    getSettingsConfiguration,
    getTranslations,
    testEnvDatabase,
    testEnvPath,
    updateEnvConfig,
    updateSettingsConfiguration,
} from './configuration';

const mockedApi = apiBase as unknown as {
    get: jest.Mock;
    put: jest.Mock;
    post: jest.Mock;
};

const settingsRequest = {
    indexing: {
        scan_on_startup: true,
        extract_metadata: true,
        generate_previews: true,
    },
    captures: {
        save_path: '/srv/capturas',
    },
    ai: {
        image_classification: true,
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
                indexing: { ...settingsRequest.indexing, workers_enabled: true },
                language: { ...settingsRequest.language, available: ['en-US', 'pt-BR'] },
            },
        },
        {
            name: 'gets env config',
            fn: () => getEnvConfig(),
            method: 'get' as const,
            url: '/configuration/env',
            payload: undefined,
            response: {
                fields: [
                    {
                        key: 'LANGUAGE',
                        group: 'general',
                        kind: 'string',
                        value: 'pt-BR',
                        configured: true,
                        dangerous: false,
                    },
                ],
                restart_required: false,
            },
        },
        {
            name: 'updates env config',
            fn: () => updateEnvConfig({ changes: { LANGUAGE: 'en-US' }, confirmed: false }),
            method: 'put' as const,
            url: '/configuration/env',
            payload: { changes: { LANGUAGE: 'en-US' }, confirmed: false },
            response: { fields: [], restart_required: true },
        },
        {
            name: 'tests env database',
            fn: () =>
                testEnvDatabase({
                    host: 'localhost',
                    port: '5432',
                    user: 'postgres',
                    name: 'kuranas',
                    password: '',
                }),
            method: 'post' as const,
            url: '/configuration/env/test-db',
            payload: {
                host: 'localhost',
                port: '5432',
                user: 'postgres',
                name: 'kuranas',
                password: '',
            },
            response: { ok: true, message: 'ok' },
        },
        {
            name: 'tests env path',
            fn: () => testEnvPath({ path: '/data' }),
            method: 'post' as const,
            url: '/configuration/env/test-path',
            payload: { path: '/data' },
            response: { ok: false, message: 'fail' },
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
