import { useCallback, useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { cancelAutoShutdownToday, getAutoShutdownStatus } from '@/service/autoShutdown';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';

export const autoShutdownStatusQueryKey = ['auto-shutdown-status'];

const idlePollingIntervalMs = 30_000;
const warningPollingIntervalMs = 5_000;
const countdownTickMs = 1_000;

export const useAutoShutdownBanner = () => {
    const queryClient = useQueryClient();
    const [nowMs, setNowMs] = useState(() => Date.now());

    const statusQuery = useQuery({
        queryKey: autoShutdownStatusQueryKey,
        queryFn: getAutoShutdownStatus,
        retry: false,
        refetchOnWindowFocus: false,
        refetchInterval: (query) =>
            query.state.data?.state === 'warning'
                ? warningPollingIntervalMs
                : idlePollingIntervalMs,
    });

    const status = statusQuery.data;
    const secondsRemaining = status?.seconds_remaining;
    const isWarning = status?.state === 'warning' && typeof secondsRemaining === 'number';

    useEffect(() => {
        if (!isWarning) return;
        const timer = setInterval(() => setNowMs(Date.now()), countdownTickMs);
        return () => clearInterval(timer);
    }, [isWarning]);

    const elapsedSeconds = Math.max(
        0,
        Math.floor((nowMs - statusQuery.dataUpdatedAt) / countdownTickMs)
    );
    const secondsLeft = isWarning ? Math.max(0, secondsRemaining - elapsedSeconds) : 0;

    const cancelMutation = useMutation({
        mutationFn: cancelAutoShutdownToday,
        onSuccess: (idleStatus) => {
            queryClient.setQueryData(autoShutdownStatusQueryKey, idleStatus);
            void queryClient.invalidateQueries({ queryKey: autoShutdownStatusQueryKey });
        },
    });

    const cancelToday = useCallback(() => cancelMutation.mutate(), [cancelMutation]);

    return {
        isWarning,
        secondsLeft,
        isCancelling: cancelMutation.isPending,
        cancelErrorMessage: extractBackendErrorMessage(cancelMutation.error) ?? null,
        cancelToday,
    };
};
