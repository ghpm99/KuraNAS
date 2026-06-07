import { apiBase } from '@/service';
import { getApiBaseUrl } from '@/service/apiUrl';
import type { DownloadItem } from '@/types/downloads';

export const getDownloads = async (): Promise<DownloadItem[]> => {
    const response = await apiBase.get<DownloadItem[]>('/downloads');
    return response.data;
};

// buildDownloadHref turns the relative download_url returned by the API into an
// absolute href the browser can navigate to directly. It is same-origin when
// the UI is served by the backend and points at the remote server in dev.
export const buildDownloadHref = (item: DownloadItem): string =>
    `${getApiBaseUrl()}${item.download_url}`;
