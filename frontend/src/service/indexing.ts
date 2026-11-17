import { apiBase } from '@/service';

export type ReindexScope =
    | { scope: 'all' }
    | { scope: 'root'; rootId: number }
    | { scope: 'folder'; folderId: number };

export type ReindexRequestResult = {
    jobId: number;
    isAlreadyRunning: boolean;
};

type ReindexJobResponse = { job_id: number };

const HTTP_STATUS_CONFLICT = 409;

const buildReindexPayload = (reindexScope: ReindexScope) => {
    switch (reindexScope.scope) {
        case 'root':
            return { scope: 'root', root_id: reindexScope.rootId };
        case 'folder':
            return { scope: 'folder', folder_id: reindexScope.folderId };
        default:
            return { scope: 'all' };
    }
};

const extractRunningJobId = (error: unknown): number | undefined => {
    const response = (
        error as { response?: { status?: number; data?: Partial<ReindexJobResponse> } } | null
    )?.response;
    if (response?.status !== HTTP_STATUS_CONFLICT) return undefined;
    return typeof response.data?.job_id === 'number' ? response.data.job_id : undefined;
};

export const requestReindex = async (
    reindexScope: ReindexScope
): Promise<ReindexRequestResult> => {
    try {
        const response = await apiBase.post<ReindexJobResponse>(
            '/indexing/reindex',
            buildReindexPayload(reindexScope)
        );
        return { jobId: response.data.job_id, isAlreadyRunning: false };
    } catch (error) {
        const runningJobId = extractRunningJobId(error);
        if (runningJobId === undefined) throw error;
        return { jobId: runningJobId, isAlreadyRunning: true };
    }
};
