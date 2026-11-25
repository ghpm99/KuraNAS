jest.mock('./index', () => ({
    apiBase: {
        get: jest.fn(),
        post: jest.fn(),
    },
}));

import { apiBase } from './index';
import {
    fetchFromUrl,
    getIngestDownloads,
    getIngestPresets,
    getIngestTargets,
    getYtDlpStatus,
    retryIngestDownload,
    updateYtDlp,
} from './ingest';

const mockedApi = apiBase as unknown as {
    get: jest.Mock;
    post: jest.Mock;
};

describe('service/ingest', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('gets the yt-dlp status', async () => {
        const status = {
            installed: true,
            current_version: '2024.08.06',
            latest_version: '2024.09.01',
            update_available: true,
            release_url: 'http://x',
            release_date: '2024-09-01',
        };
        mockedApi.get.mockResolvedValue({ data: status });

        const result = await getYtDlpStatus();

        expect(mockedApi.get).toHaveBeenCalledWith('/ingest/ytdlp/status');
        expect(result).toEqual(status);
    });

    it('triggers a yt-dlp update', async () => {
        mockedApi.post.mockResolvedValue({ data: { message: 'ok' } });

        await updateYtDlp();

        expect(mockedApi.post).toHaveBeenCalledWith('/ingest/ytdlp/update');
    });

    it('requests a download with the backend body and decodes job and download ids', async () => {
        mockedApi.post.mockResolvedValue({ data: { job_id: 345, download_id: 12 } });
        const request = {
            url: 'https://exemplo.com/video',
            preset: 'video_1080',
            target_root: '/media/videos',
            subfolder: 'series',
        };

        const response = await fetchFromUrl(request);

        expect(mockedApi.post).toHaveBeenCalledWith('/ingest/fetch', request);
        expect(response).toEqual({ job_id: 345, download_id: 12 });
    });

    it('lists the download targets', async () => {
        const targets = [{ label: 'Videos', path: '/media/videos' }];
        mockedApi.get.mockResolvedValue({ data: targets });

        const response = await getIngestTargets();

        expect(mockedApi.get).toHaveBeenCalledWith('/ingest/targets');
        expect(response).toEqual(targets);
    });

    it('lists the download presets', async () => {
        const presets = [
            { key: 'audio_mp3', label: 'Audio MP3' },
            { key: 'video_1080', label: 'Video 1080p' },
        ];
        mockedApi.get.mockResolvedValue({ data: presets });

        const response = await getIngestPresets();

        expect(mockedApi.get).toHaveBeenCalledWith('/ingest/presets');
        expect(response).toEqual(presets);
    });

    const runningDownload = {
        id: 12,
        job_id: 345,
        url: 'https://exemplo.com/video',
        title: null,
        preset: 'video_1080',
        target_root_id: 3,
        target_root_name: 'Videos',
        subfolder: 'series',
        state: 'running',
        progress: 42,
        speed_bytes_per_second: 1048576,
        eta_seconds: 30,
        file_id: null,
        error: '',
        created_at: '2026-11-25T10:00:00Z',
        ended_at: null,
    };

    it('lists downloads sending page, page_size and state', async () => {
        const envelope = {
            items: [runningDownload],
            pagination: { page: 2, page_size: 20, has_next: true, has_prev: true },
        };
        mockedApi.get.mockResolvedValue({ data: envelope });

        const response = await getIngestDownloads({ page: 2, pageSize: 20, state: 'active' });

        expect(mockedApi.get).toHaveBeenCalledWith('/ingest/downloads', {
            params: { page: 2, page_size: 20, state: 'active' },
        });
        expect(response.items[0]?.job_id).toBe(345);
        expect(response.pagination.has_next).toBe(true);
    });

    it('omits the state filter when none is given', async () => {
        mockedApi.get.mockResolvedValue({
            data: {
                items: [],
                pagination: { page: 1, page_size: 20, has_next: false, has_prev: false },
            },
        });

        await getIngestDownloads({ page: 1, pageSize: 20 });

        expect(mockedApi.get).toHaveBeenCalledWith('/ingest/downloads', {
            params: { page: 1, page_size: 20 },
        });
    });

    it('retries a download by id and decodes the new item', async () => {
        const retriedDownload = { ...runningDownload, id: 13, state: 'queued', progress: 0 };
        mockedApi.post.mockResolvedValue({ data: retriedDownload });

        const response = await retryIngestDownload(12);

        expect(mockedApi.post).toHaveBeenCalledWith('/ingest/downloads/12/retry');
        expect(response).toEqual(retriedDownload);
    });
});
