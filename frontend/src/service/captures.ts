import type { Pagination } from '@/types/pagination';
import type { Capture } from '@/types/captures';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import { apiBase } from './index';

interface ListCapturesParams {
    page?: number;
    pageSize?: number;
    name?: string;
    mediaType?: string;
}

export const getCaptures = async ({
    page = 1,
    pageSize = 50,
    name,
    mediaType,
}: ListCapturesParams = {}): Promise<Pagination<Capture>> => {
    const response = await apiBase.get<Pagination<Capture>>('/captures', {
        params: {
            page,
            page_size: pageSize,
            name,
            media_type: mediaType,
        },
    });
    return response.data;
};

export const getCaptureById = async (id: number): Promise<Capture> => {
    const response = await apiBase.get<Capture>(`/captures/${id}`);
    return response.data;
};

export const deleteCapture = async (id: number): Promise<void> => {
    await apiBase.delete(`/captures/${id}`);
};

// captureThumbnailHref prefers the provenance thumbnail captured from the source
// site; once the capture is promoted it falls back to the video thumbnail of the
// final file. Returns an empty string when neither is available.
export const captureThumbnailHref = (capture: Capture): string => {
    if (capture.thumbnail_url) {
        return capture.thumbnail_url;
    }
    if (capture.status === 'promoted' && capture.file_id) {
        return `${getApiV1BaseUrl()}/files/video-thumbnail/${capture.file_id}?width=320&height=180`;
    }
    return '';
};
