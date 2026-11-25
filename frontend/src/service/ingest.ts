import { apiBase } from '@/service';
import type {
    IngestDownload,
    IngestDownloadsPageParams,
    IngestFetchRequest,
    IngestFetchResponse,
    IngestPreset,
    IngestTarget,
    YtDlpStatus,
} from '@/types/ingest';
import type { Pagination } from '@/types/pagination';

export const getYtDlpStatus = async (): Promise<YtDlpStatus> => {
    const response = await apiBase.get<YtDlpStatus>('/ingest/ytdlp/status');
    return response.data;
};

export const updateYtDlp = async (): Promise<void> => {
    await apiBase.post('/ingest/ytdlp/update');
};

export const fetchFromUrl = async (request: IngestFetchRequest): Promise<IngestFetchResponse> => {
    const response = await apiBase.post<IngestFetchResponse>('/ingest/fetch', request);
    return response.data;
};

export const getIngestTargets = async (): Promise<IngestTarget[]> => {
    const response = await apiBase.get<IngestTarget[]>('/ingest/targets');
    return response.data;
};

export const getIngestPresets = async (): Promise<IngestPreset[]> => {
    const response = await apiBase.get<IngestPreset[]>('/ingest/presets');
    return response.data;
};

export const getIngestDownloads = async ({
    page,
    pageSize,
    state,
}: IngestDownloadsPageParams): Promise<Pagination<IngestDownload>> => {
    const response = await apiBase.get<Pagination<IngestDownload>>('/ingest/downloads', {
        params: { page, page_size: pageSize, ...(state ? { state } : {}) },
    });
    return response.data;
};

export const retryIngestDownload = async (downloadId: number): Promise<IngestDownload> => {
    const response = await apiBase.post<IngestDownload>(`/ingest/downloads/${downloadId}/retry`);
    return response.data;
};
