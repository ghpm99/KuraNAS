export type YtDlpStatus = {
    installed: boolean;
    current_version: string;
    latest_version: string;
    update_available: boolean;
    release_url: string;
    release_date: string;
};

export type IngestFetchRequest = {
    url: string;
    preset: string;
    target_root: string;
    subfolder: string;
};

export type IngestFetchResponse = {
    job_id: number;
    download_id: number;
};

export type IngestTarget = {
    label: string;
    path: string;
};

export type IngestPreset = {
    key: string;
    label: string;
};

export type IngestDownloadStateFilter = 'active' | 'finished';

export type IngestDownload = {
    id: number;
    job_id: number | null;
    url: string;
    title: string | null;
    preset: string;
    target_root_id: number;
    target_root_name: string;
    subfolder: string;
    state: string;
    progress: number;
    speed_bytes_per_second: number | null;
    eta_seconds: number | null;
    file_id: number | null;
    error: string;
    created_at: string;
    ended_at: string | null;
};

export type IngestDownloadsPageParams = {
    page: number;
    pageSize: number;
    state?: IngestDownloadStateFilter;
};
