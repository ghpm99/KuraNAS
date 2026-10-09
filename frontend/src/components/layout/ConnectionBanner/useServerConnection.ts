import { useCallback, useEffect, useState, useSyncExternalStore } from 'react';
import { onlineManager, useQueryClient } from '@tanstack/react-query';
import { getServerHealth } from '@/service/health';
import { isNetworkError } from '@/service/networkError';
import { getReconnectDelay } from './reconnectBackoff';

const subscribeToBrowserConnectivity = (onChange: () => void) => onlineManager.subscribe(onChange);
const readBrowserConnectivity = () => onlineManager.isOnline();

const isFailedQueryAction = (action: { type: string; error?: unknown }) =>
    (action.type === 'failed' || action.type === 'error') && isNetworkError(action.error);

export const useServerConnection = () => {
    const queryClient = useQueryClient();
    const isBrowserOnline = useSyncExternalStore(
        subscribeToBrowserConnectivity,
        readBrowserConnectivity
    );
    const [hasNetworkFailure, setHasNetworkFailure] = useState(false);
    const isUnavailable = !isBrowserOnline || hasNetworkFailure;

    useEffect(
        () =>
            queryClient.getQueryCache().subscribe((event) => {
                if (event.type !== 'updated') return;
                if (isFailedQueryAction(event.action)) setHasNetworkFailure(true);
            }),
        [queryClient]
    );

    const checkServerHealth = useCallback(async (): Promise<boolean> => {
        try {
            await getServerHealth();
        } catch {
            return false;
        }
        setHasNetworkFailure(false);
        await queryClient.invalidateQueries();
        return true;
    }, [queryClient]);

    useEffect(() => {
        if (!isUnavailable) return;

        let attemptIndex = 0;
        let isCancelled = false;
        let timerId: ReturnType<typeof setTimeout>;

        const scheduleNextCheck = () => {
            timerId = setTimeout(async () => {
                const isRecovered = await checkServerHealth();
                if (isCancelled || isRecovered) return;
                attemptIndex += 1;
                scheduleNextCheck();
            }, getReconnectDelay(attemptIndex));
        };

        scheduleNextCheck();

        return () => {
            isCancelled = true;
            clearTimeout(timerId);
        };
    }, [isUnavailable, checkServerHealth]);

    return { isUnavailable, retryNow: checkServerHealth };
};
