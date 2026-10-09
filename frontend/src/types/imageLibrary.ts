export const imageCategories = [
    'capture',
    'photo',
    'other',
    'document',
    'receipt',
    'landscape',
    'portrait',
    'meme',
    'art',
    'screenshot_app',
] as const;

export type ImageCategory = (typeof imageCategories)[number];

export const imageFormats = [
    'jpg',
    'jpeg',
    'jfif',
    'png',
    'gif',
    'bmp',
    'svg',
    'webp',
    'tif',
    'tiff',
    'heic',
    'heif',
    'avif',
    'cr2',
    'cr3',
    'nef',
    'arw',
    'dng',
    'orf',
    'rw2',
    'raf',
    'srw',
    'pef',
] as const;

export type ImageFormat = (typeof imageFormats)[number];

export type ImageLibrarySort = 'taken_at' | 'name' | 'size';
export type ImageLibrarySortOrder = 'asc' | 'desc';

export type ImageLibraryOrdering = {
    sort: ImageLibrarySort;
    order: ImageLibrarySortOrder;
};

export type ImageLibraryFilters = {
    nameQuery: string;
    categories: ImageCategory[];
    isStarredOnly: boolean;
    formats: string[];
    camera: string;
    takenFrom: string;
    takenTo: string;
    folder: string;
};

export type ImageLibraryItem = {
    file_id: number;
    name: string;
    path: string;
    parent_path: string;
    format: string;
    size: number;
    width: number;
    height: number;
    taken_at: string | null;
    category: ImageCategory;
    starred: boolean;
    tier: string;
    updated_at: string;
};

export type ImageLibraryPage = {
    items: ImageLibraryItem[];
    next_cursor: string;
    has_next: boolean;
    page_size: number;
    page?: number;
};

export type ImageLibraryCount = { total: number };

export type ImageTimelineBucket = {
    year: number;
    month: number;
    count: number;
};

export type ImageCameraFacet = {
    camera: string;
    count: number;
};

export type ImageFormatFacet = {
    format: string;
    count: number;
};

export type ImageMetadataSummary = {
    width: number;
    height: number;
    make: string;
    model: string;
    lens_model: string;
    datetime_original: string;
    exposure_time: number;
    f_number: number;
    iso: number;
    focal_length: number;
    software?: string;
    image_description?: string;
    taken_at?: string | null;
    gps_latitude?: number | null;
    gps_longitude?: number | null;
    classification_confidence?: number;
    suggested_name?: string;
    caption?: string;
    tags?: string[];
    ocr_text?: string;
};

export type ImageLibraryNeighbors = {
    before: ImageLibraryItem[];
    after: ImageLibraryItem[];
};

export type ImageLibraryFolder = {
    path: string;
    name: string;
    image_count: number;
    cover_file_id: number;
};
