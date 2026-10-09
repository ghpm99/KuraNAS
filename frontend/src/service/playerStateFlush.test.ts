import { flushPlayerQueue, flushPlayerState } from './playerStateFlush';
import { getPlayerClientId } from './playerClientId';

const readBlobAsText = (blob: Blob) =>
    new Promise<string>((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => resolve(String(reader.result));
        reader.onerror = () => reject(reader.error);
        reader.readAsText(blob);
    });

describe('service/playerStateFlush', () => {
    const originalSendBeacon = navigator.sendBeacon;
    const originalFetch = globalThis.fetch;

    afterEach(() => {
        Object.defineProperty(navigator, 'sendBeacon', {
            value: originalSendBeacon,
            configurable: true,
        });
        globalThis.fetch = originalFetch;
    });

    const stubSendBeacon = (implementation: unknown) =>
        Object.defineProperty(navigator, 'sendBeacon', {
            value: implementation,
            configurable: true,
        });

    it('sends the queue with sendBeacon carrying the client id in the query string', async () => {
        const sendBeacon = jest.fn().mockReturnValue(true);
        stubSendBeacon(sendBeacon);
        globalThis.fetch = jest.fn();

        flushPlayerQueue({ file_ids: [1, 2], current_index: 1 });

        const [url, body] = sendBeacon.mock.calls[0];
        expect(url).toBe(
            `/api/v1/music/player-state/queue?client_id=${encodeURIComponent(getPlayerClientId())}`
        );
        expect(JSON.parse(await readBlobAsText(body as Blob))).toEqual({
            file_ids: [1, 2],
            current_index: 1,
        });
        expect(globalThis.fetch).not.toHaveBeenCalled();
    });

    it('sends the state with sendBeacon to the player-state endpoint', async () => {
        const sendBeacon = jest.fn().mockReturnValue(true);
        stubSendBeacon(sendBeacon);

        flushPlayerState({ current_file_id: 4, current_position: 9 });

        expect(sendBeacon.mock.calls[0][0]).toContain('/music/player-state/?client_id=');
        expect(JSON.parse(await readBlobAsText(sendBeacon.mock.calls[0][1] as Blob))).toEqual({
            current_file_id: 4,
            current_position: 9,
        });
    });

    it.each([
        ['sendBeacon refuses the payload', jest.fn().mockReturnValue(false)],
        ['sendBeacon is unavailable', undefined],
    ])('falls back to fetch keepalive when %s', (_scenario, sendBeacon) => {
        stubSendBeacon(sendBeacon);
        const fetchMock = jest.fn().mockResolvedValue({});
        globalThis.fetch = fetchMock;

        flushPlayerQueue({ file_ids: [3], current_index: 0 });

        const [url, init] = fetchMock.mock.calls[0];
        expect(url).toContain('/music/player-state/queue?client_id=');
        expect(init).toMatchObject({
            method: 'PUT',
            keepalive: true,
            headers: { 'X-KuraNAS-Client-Id': getPlayerClientId() },
            body: JSON.stringify({ file_ids: [3], current_index: 0 }),
        });
    });

    it('swallows a rejected keepalive fetch', async () => {
        stubSendBeacon(undefined);
        globalThis.fetch = jest.fn().mockRejectedValue(new Error('offline'));

        expect(() => flushPlayerState({ current_position: 1 })).not.toThrow();
        await Promise.resolve();
    });
});
