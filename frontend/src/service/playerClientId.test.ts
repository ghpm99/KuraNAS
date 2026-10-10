import type { InternalAxiosRequestConfig } from 'axios';
import { AxiosHeaders } from 'axios';

type PlayerClientIdModule = typeof import('./playerClientId');

const loadFreshModule = (): PlayerClientIdModule => {
    let freshModule!: PlayerClientIdModule;
    jest.isolateModules(() => {
        freshModule = require('./playerClientId');
    });
    return freshModule;
};

const buildRequestConfig = (url: string) =>
    ({ url, headers: new AxiosHeaders() }) as InternalAxiosRequestConfig;

describe('service/playerClientId', () => {
    afterEach(() => {
        jest.restoreAllMocks();
        window.localStorage.clear();
    });

    it('generates a valid client id and persists it in localStorage', () => {
        const { getPlayerClientId } = loadFreshModule();

        const clientId = getPlayerClientId();

        expect(clientId).toMatch(/^[A-Za-z0-9-]{8,64}$/);
        expect(window.localStorage.getItem('kuranas.playerClientId')).toBe(clientId);
        expect(getPlayerClientId()).toBe(clientId);
    });

    it('reuses the stored client id across loads', () => {
        window.localStorage.setItem('kuranas.playerClientId', 'stored-device-0001');

        expect(loadFreshModule().getPlayerClientId()).toBe('stored-device-0001');
    });

    it('replaces a stored value that is not a valid client id', () => {
        window.localStorage.setItem('kuranas.playerClientId', 'bad value!');

        const clientId = loadFreshModule().getPlayerClientId();

        expect(clientId).not.toBe('bad value!');
        expect(clientId).toMatch(/^[A-Za-z0-9-]{8,64}$/);
    });

    it('falls back to a session only id when localStorage is unavailable', () => {
        jest.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('denied');
        });
        jest.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('denied');
        });
        const { getPlayerClientId } = loadFreshModule();

        const firstClientId = getPlayerClientId();

        expect(firstClientId).toMatch(/^[A-Za-z0-9-]{8,64}$/);
        expect(getPlayerClientId()).toBe(firstClientId);
    });

    it('generates an id without crypto.randomUUID', () => {
        const originalRandomUUID = crypto.randomUUID;
        Object.defineProperty(crypto, 'randomUUID', { value: undefined, configurable: true });
        try {
            expect(loadFreshModule().getPlayerClientId()).toMatch(/^[A-Za-z0-9-]{8,64}$/);
        } finally {
            Object.defineProperty(crypto, 'randomUUID', {
                value: originalRandomUUID,
                configurable: true,
            });
        }
    });

    it('attaches the header only to music and video requests', () => {
        const { attachPlayerClientId, PLAYER_CLIENT_ID_HEADER, getPlayerClientId } =
            loadFreshModule();

        const musicConfig = attachPlayerClientId(buildRequestConfig('/music/player-state/'));
        const videoConfig = attachPlayerClientId(buildRequestConfig('/video/playback/state'));
        const filesConfig = attachPlayerClientId(buildRequestConfig('/files/'));
        const urllessConfig = attachPlayerClientId({
            headers: new AxiosHeaders(),
        } as InternalAxiosRequestConfig);

        expect(musicConfig.headers.get(PLAYER_CLIENT_ID_HEADER)).toBe(getPlayerClientId());
        expect(videoConfig.headers.get(PLAYER_CLIENT_ID_HEADER)).toBe(getPlayerClientId());
        expect(filesConfig.headers.has(PLAYER_CLIENT_ID_HEADER)).toBe(false);
        expect(urllessConfig.headers.has(PLAYER_CLIENT_ID_HEADER)).toBe(false);
    });
});
