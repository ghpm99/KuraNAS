export type DownloadItem = {
    id: string;
    name: string;
    description: string;
    platform: string;
    version: string;
    min_os: string;
    size_bytes: number;
    sha256: string;
    download_url: string;
};
