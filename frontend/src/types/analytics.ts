export type AnalyticsPeriod = '24h' | '7d' | '30d' | '90d';

export interface AnalyticsOverview {
    period: AnalyticsPeriod;
    generated_at: string;
    storage: {
        total_bytes: number;
        used_bytes: number;
        free_bytes: number;
        growth_bytes: number;
    };
    counts: {
        files_total: number;
        files_added: number;
        folders: number;
    };
    time_series: Array<{
        date: string;
        used_bytes: number;
    }>;
    types: Array<{
        type: string;
        count: number;
        bytes: number;
    }>;
    extensions: Array<{
        ext: string;
        count: number;
        bytes: number;
    }>;
    hot_folders: Array<{
        path: string;
        new_files: number;
        added_bytes: number;
        last_event: string;
    }>;
    top_folders: Array<{
        path: string;
        files: number;
        bytes: number;
        last_modified: string;
    }>;
    recent_files: Array<{
        id: number;
        name: string;
        path: string;
        parent_path: string;
        format: string;
        size_bytes: number;
        created_at: string;
        updated_at: string;
    }>;
    duplicates: {
        groups: number;
        files: number;
        reclaimable_size: number;
        top_groups: Array<{
            signature: string;
            copies: number;
            size_bytes: number;
            reclaimable_size: number;
            paths: string[];
        }>;
    };
    library: {
        categorized_media: number;
        audio_with_metadata: number;
        video_with_metadata: number;
        image_with_metadata: number;
        image_classified: number;
    };
    processing: {
        metadata_pending: number;
        metadata_failed: number;
        thumbnail_pending: number;
        thumbnail_failed: number;
    };
    health: {
        status: 'ok' | 'scanning' | 'error';
        last_scan_at: string;
        last_scan_seconds: number;
        indexed_files: number;
        errors_last_24h: number;
        recent_errors: string[];
    };
}
