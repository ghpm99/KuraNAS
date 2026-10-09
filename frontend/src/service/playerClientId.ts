import type { InternalAxiosRequestConfig } from 'axios';

export const PLAYER_CLIENT_ID_HEADER = 'X-KuraNAS-Client-Id';

const PLAYER_CLIENT_ID_STORAGE_KEY = 'kuranas.playerClientId';
const PLAYER_CLIENT_ID_PATTERN = /^[A-Za-z0-9-]{8,64}$/;
const MUSIC_API_PATH_PREFIX = '/music/';

let sessionClientId: string | undefined;

const generateClientId = (): string => {
    if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
        return crypto.randomUUID();
    }
    const randomSegment = () => Math.random().toString(16).slice(2, 10).padEnd(8, '0');
    return `${randomSegment()}-${randomSegment()}-${randomSegment()}`;
};

const readStoredClientId = (): string | undefined => {
    try {
        const storedClientId = window.localStorage.getItem(PLAYER_CLIENT_ID_STORAGE_KEY);
        return storedClientId && PLAYER_CLIENT_ID_PATTERN.test(storedClientId)
            ? storedClientId
            : undefined;
    } catch {
        return undefined;
    }
};

const storeClientId = (clientId: string) => {
    try {
        window.localStorage.setItem(PLAYER_CLIENT_ID_STORAGE_KEY, clientId);
    } catch {
        sessionClientId = clientId;
    }
};

export const getPlayerClientId = (): string => {
    const storedClientId = readStoredClientId();
    if (storedClientId) {
        return storedClientId;
    }
    if (sessionClientId) {
        return sessionClientId;
    }
    const generatedClientId = generateClientId();
    sessionClientId = generatedClientId;
    storeClientId(generatedClientId);
    return generatedClientId;
};

export const attachPlayerClientId = (
    requestConfig: InternalAxiosRequestConfig
): InternalAxiosRequestConfig => {
    if (requestConfig.url?.startsWith(MUSIC_API_PATH_PREFIX)) {
        requestConfig.headers.set(PLAYER_CLIENT_ID_HEADER, getPlayerClientId());
    }
    return requestConfig;
};
