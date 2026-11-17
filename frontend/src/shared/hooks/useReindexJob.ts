import { useCallback, useEffect, useRef, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import { requestReindex, type ReindexScope } from '@/service/indexing';
import { getJob } from '@/service/jobs';
import { isJobFinished, type Job } from '@/types/jobs';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';

const JOB_POLLING_INTERVAL_MS = 2000;

type UseReindexJobOptions = {
    onFinished?: (job: Job) => void;
};

const useReindexJob = ({ onFinished }: UseReindexJobOptions = {}) => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const [trackedJobId, setTrackedJobId] = useState<number | null>(null);
    const [isRequesting, setIsRequesting] = useState(false);
    const [errorMessage, setErrorMessage] = useState<string | null>(null);
    const finishedJobIdRef = useRef<number | null>(null);
    const onFinishedRef = useRef(onFinished);

    useEffect(() => {
        onFinishedRef.current = onFinished;
    }, [onFinished]);

    const { data: job } = useQuery({
        queryKey: ['reindex-job', trackedJobId],
        queryFn: () => getJob(trackedJobId as number),
        enabled: trackedJobId !== null,
        refetchInterval: (query) =>
            query.state.data && isJobFinished(query.state.data.status)
                ? false
                : JOB_POLLING_INTERVAL_MS,
    });

    const hasFinished = job !== undefined && isJobFinished(job.status);

    useEffect(() => {
        if (!job || !hasFinished || finishedJobIdRef.current === job.id) return;
        finishedJobIdRef.current = job.id;
        if (job.status === 'completed') {
            enqueueSnackbar(t('REINDEX_FINISHED'), { variant: 'success' });
        } else {
            enqueueSnackbar(job.last_error || t('REINDEX_FAILED'), { variant: 'error' });
        }
        onFinishedRef.current?.(job);
    }, [job, hasFinished, enqueueSnackbar, t]);

    const start = useCallback(
        async (reindexScope: ReindexScope) => {
            setIsRequesting(true);
            setErrorMessage(null);
            try {
                const { jobId, isAlreadyRunning } = await requestReindex(reindexScope);
                finishedJobIdRef.current = null;
                setTrackedJobId(jobId);
                enqueueSnackbar(
                    t(isAlreadyRunning ? 'REINDEX_ALREADY_RUNNING' : 'REINDEX_STARTED'),
                    { variant: 'info' }
                );
            } catch (error) {
                const message = extractBackendErrorMessage(error) ?? t('REINDEX_ERROR_FAILED');
                setErrorMessage(message);
                enqueueSnackbar(message, { variant: 'error' });
            } finally {
                setIsRequesting(false);
            }
        },
        [enqueueSnackbar, t]
    );

    const isTracking = trackedJobId !== null && !hasFinished;

    return {
        start,
        isRunning: isRequesting || isTracking,
        progress: job?.progress?.progress ?? 0,
        job: job ?? null,
        errorMessage,
    };
};

export default useReindexJob;
