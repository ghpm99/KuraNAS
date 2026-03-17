import { viteEnv } from '@/config/viteEnv';

const normalizeBase = (value: string): string => value.replace(/\/+$/, '');
export const getApiBaseUrl = (): string => {
    const runtimeGlobal = (globalThis as { __KURANAS_API_URL__?: string }).__KURANAS_API_URL__;
    if (typeof runtimeGlobal === 'string' && runtimeGlobal.trim().length > 0) {
        return normalizeBase(runtimeGlobal);
    }

    const viteEnvApiUrl = viteEnv?.VITE_API_URL;
    if (typeof viteEnvApiUrl === 'string' && viteEnvApiUrl.trim().length > 0) {
        return normalizeBase(viteEnvApiUrl);
    }

    // Fallback for non-Vite runtimes (e.g. Jest/Node).
    const envApiUrl = typeof process !== 'undefined' ? process.env.VITE_API_URL : undefined;
    if (typeof envApiUrl === 'string' && envApiUrl.trim().length > 0) {
        return normalizeBase(envApiUrl);
    }

    return '';
};

export const getApiV1BaseUrl = (): string => {
    const base = getApiBaseUrl();
    return base ? `${base}/api/v1` : '/api/v1';
};
