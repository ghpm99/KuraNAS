import { getApiV1BaseUrl } from './apiUrl';
import { getPlayerClientId, PLAYER_CLIENT_ID_HEADER } from './playerClientId';
import type { ReplacePlayerQueueRequest, UpdatePlayerStateRequest } from './playerState';

const PLAYER_STATE_PATH = '/music/player-state/';
const PLAYER_QUEUE_PATH = '/music/player-state/queue';
const JSON_CONTENT_TYPE = 'application/json';

const flushJson = (path: string, body: object) => {
    const clientId = getPlayerClientId();
    const url = `${getApiV1BaseUrl()}${path}?client_id=${encodeURIComponent(clientId)}`;
    const payload = JSON.stringify(body);

    const wasQueuedByBeacon =
        typeof navigator.sendBeacon === 'function' &&
        navigator.sendBeacon(url, new Blob([payload], { type: JSON_CONTENT_TYPE }));
    if (wasQueuedByBeacon) {
        return;
    }

    fetch(url, {
        method: 'PUT',
        keepalive: true,
        headers: { 'Content-Type': JSON_CONTENT_TYPE, [PLAYER_CLIENT_ID_HEADER]: clientId },
        body: payload,
    }).catch(() => undefined);
};

export const flushPlayerState = (state: UpdatePlayerStateRequest) =>
    flushJson(PLAYER_STATE_PATH, state);

export const flushPlayerQueue = (queue: ReplacePlayerQueueRequest) =>
    flushJson(PLAYER_QUEUE_PATH, queue);
