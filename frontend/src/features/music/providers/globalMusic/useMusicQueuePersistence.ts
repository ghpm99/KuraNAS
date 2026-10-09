import { useEffect, useRef } from 'react';
import type { IMusicData } from '../musicProvider/musicProvider';
import { replacePlayerQueue, type ReplacePlayerQueueRequest } from '@/service/playerState';
import { flushPlayerQueue } from '@/service/playerStateFlush';

export const QUEUE_PERSIST_DEBOUNCE_MS = 2000;
export const MAX_PERSISTED_QUEUE_ENTRIES = 10000;

type PersistenceParams = {
    isEnabled: boolean;
    queue: IMusicData[];
    currentIndex: number | undefined;
};

const buildQueueRequest = (
    queue: IMusicData[],
    currentIndex: number | undefined
): ReplacePlayerQueueRequest => {
    const persistedIds = queue.slice(0, MAX_PERSISTED_QUEUE_ENTRIES).map((track) => track.id);
    const lastValidIndex = Math.max(persistedIds.length - 1, 0);
    return {
        file_ids: persistedIds,
        current_index: Math.min(currentIndex ?? 0, lastValidIndex),
    };
};

export default function useMusicQueuePersistence({
    isEnabled,
    queue,
    currentIndex,
}: PersistenceParams) {
    const lastSignatureRef = useRef<string | null>(null);
    const pendingRequestRef = useRef<ReplacePlayerQueueRequest | null>(null);
    const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    useEffect(() => {
        if (!isEnabled) {
            lastSignatureRef.current = null;
            return;
        }

        const request = buildQueueRequest(queue, currentIndex);
        const signature = JSON.stringify(request);
        if (lastSignatureRef.current === null) {
            lastSignatureRef.current = signature;
            return;
        }
        if (lastSignatureRef.current === signature) {
            return;
        }
        lastSignatureRef.current = signature;

        if (timeoutRef.current) clearTimeout(timeoutRef.current);
        pendingRequestRef.current = request;
        timeoutRef.current = setTimeout(() => {
            pendingRequestRef.current = null;
            replacePlayerQueue(request).catch(() => undefined);
        }, QUEUE_PERSIST_DEBOUNCE_MS);
    }, [isEnabled, queue, currentIndex]);

    useEffect(() => {
        const flushPendingQueue = () => {
            const pendingRequest = pendingRequestRef.current;
            if (!pendingRequest) return;
            if (timeoutRef.current) clearTimeout(timeoutRef.current);
            pendingRequestRef.current = null;
            flushPlayerQueue(pendingRequest);
        };

        window.addEventListener('pagehide', flushPendingQueue);
        return () => {
            window.removeEventListener('pagehide', flushPendingQueue);
            if (timeoutRef.current) clearTimeout(timeoutRef.current);
        };
    }, []);
}
