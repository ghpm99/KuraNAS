export type AnalyticsPeriod = '24h' | '7d' | '30d' | '90d';

export interface StorageStats {
    total_bytes: number;
    used_bytes: number;
    free_bytes: number;
    growth_bytes: number;
}

export interface CountStats {
    files_total: number;
    files_added: number;
    folders: number;
}

export interface StorageStatsResponse {
    storage: StorageStats;
    counts: CountStats;
}

export interface TimeSeriesPoint {
    date: string;
    used_bytes: number;
}

export interface TypeBreakdown {
    type: string;
    count: number;
    bytes: number;
}

export interface ExtensionStat {
    ext: string;
    count: number;
    bytes: number;
}

export interface HotFolder {
    path: string;
    new_files: number;
    added_bytes: number;
    last_event: string;
}

export interface FolderUsage {
    path: string;
    files: number;
    bytes: number;
    last_modified: string;
}

export interface RecentFile {
    id: number;
    name: string;
    path: string;
    parent_path: string;
    format: string;
    size_bytes: number;
    created_at: string;
    updated_at: string;
}

export interface DuplicatesSummary {
    groups: number;
    files: number;
    reclaimable_size: number;
}

export interface DuplicateGroup {
    signature: string;
    copies: number;
    size_bytes: number;
    reclaimable_size: number;
    paths: string[];
}

export interface LibraryStats {
    categorized_media: number;
    audio_with_metadata: number;
    video_with_metadata: number;
    image_with_metadata: number;
    image_classified: number;
}

export interface ProcessingStats {
    metadata_pending: number;
    metadata_failed: number;
    thumbnail_pending: number;
    thumbnail_failed: number;
    recurring_timeouts: number;
}

export interface HealthStatus {
    status: 'ok' | 'scanning' | 'error';
    last_scan_at: string;
    last_scan_seconds: number;
    indexed_files: number;
    errors_last_24h: number;
    recent_errors: string[];
}

export interface AnalyticsDuplicates extends DuplicatesSummary {
    top_groups: DuplicateGroup[];
}

export interface AnalyticsOverview {
    period: AnalyticsPeriod;
    generated_at: string;
    storage: StorageStats;
    counts: CountStats;
    time_series: TimeSeriesPoint[];
    types: TypeBreakdown[];
    extensions: ExtensionStat[];
    hot_folders: HotFolder[];
    top_folders: FolderUsage[];
    recent_files: RecentFile[];
    duplicates: AnalyticsDuplicates;
    library: LibraryStats;
    processing: ProcessingStats;
    health: HealthStatus;
}
