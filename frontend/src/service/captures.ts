import type { Pagination } from '@/types/pagination';
import type { Capture, CaptureStatus } from '@/types/captures';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import { apiBase } from './index';

interface ListCapturesParams {
    page?: number;
    pageSize?: number;
    name?: string;
    mediaType?: string;
    status?: CaptureStatus;
    platform?: string;
}

export const getCaptures = async ({
    page = 1,
    pageSize = 50,
    name,
    mediaType,
    status,
    platform,
}: ListCapturesParams = {}): Promise<Pagination<Capture>> => {
    const response = await apiBase.get<Pagination<Capture>>('/captures', {
        params: {
            page,
            page_size: pageSize,
            name,
            media_type: mediaType,
            ...(status ? { status } : {}),
            ...(platform ? { platform } : {}),
        },
    });
    if (!Array.isArray(response.data?.items)) {
        throw new Error('invalid captures response');
    }
    return response.data;
};

export const getCapturePlatforms = async (): Promise<string[]> => {
    const response = await apiBase.get<{ items: unknown[] }>('/captures/platforms');
    if (!Array.isArray(response.data?.items)) {
        throw new Error('invalid capture platforms response');
    }
    return response.data.items.filter(
        (platform): platform is string => typeof platform === 'string'
    );
};

export const getCaptureById = async (id: number): Promise<Capture> => {
    const response = await apiBase.get<Capture>(`/captures/${id}`);
    return response.data;
};

export type DeleteCaptureResult = {
    deleted: boolean;
    library_file_kept: boolean;
};

export const deleteCapture = async (id: number): Promise<DeleteCaptureResult> => {
    const response = await apiBase.delete<DeleteCaptureResult>(`/captures/${id}`);
    return response.data;
};

export type PromoteCaptureResult = {
    job_id: number;
};

export const promoteCapture = async (id: number): Promise<PromoteCaptureResult> => {
    const response = await apiBase.post<PromoteCaptureResult>(`/captures/${id}/promote`);
    return response.data;
};

/** Miniatura da origem; depois da promoção, a do vídeo final; string vazia quando não há nenhuma. */
export const captureThumbnailHref = (capture: Capture): string => {
    if (capture.thumbnail_url) {
        return capture.thumbnail_url;
    }
    if (capture.status === 'promoted' && capture.file_id) {
        return `${getApiV1BaseUrl()}/files/video-thumbnail/${capture.file_id}?width=320&height=180`;
    }
    return '';
};
