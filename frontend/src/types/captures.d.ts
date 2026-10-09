export type CaptureStatus = 'uploaded' | 'promoting' | 'promoted' | 'failed';

export interface Capture {
    id: number;
    name: string;
    file_name: string;
    file_path: string;
    media_type: string;
    mime_type: string;
    size: number;
    episode_key: string;
    created_at: string;
    file_id?: number;
    status: CaptureStatus;
    title?: string;
    episode_title?: string;
    season?: number;
    episode?: number;
    description?: string;
    release_year?: number;
    genres?: string[];
    cast?: string[];
    directors?: string[];
    studio?: string;
    content_rating?: string;
    platform?: string;
    source_url?: string;
    thumbnail_url?: string;
    content_type?: string;
}
