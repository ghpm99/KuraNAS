import type {
    VideoLibraryDuration,
    VideoLibraryFilters,
    VideoLibraryResolution,
} from '@/service/videoPlayback';

export const noVideoLibraryFilters: VideoLibraryFilters = { formats: [] };

export const videoLibraryFormatOptions: string[] = [
    '.mp4',
    '.mkv',
    '.avi',
    '.mov',
    '.webm',
    '.m4v',
    '.wmv',
    '.flv',
    '.ogv',
    '.mpg',
    '.mpeg',
    '.m2ts',
    '.3gp',
    '.vob',
];

export const videoLibraryResolutionOptions: VideoLibraryResolution[] = ['sd', 'hd', 'fhd', 'uhd'];

export const videoLibraryDurationOptions: VideoLibraryDuration[] = ['short', 'medium', 'long'];

export const hasActiveVideoLibraryFilters = (filters: VideoLibraryFilters): boolean =>
    filters.formats.length > 0 ||
    filters.resolution !== undefined ||
    filters.duration !== undefined;

export const toggleVideoLibraryFormat = (
    filters: VideoLibraryFilters,
    format: string
): VideoLibraryFilters => ({
    ...filters,
    formats: filters.formats.includes(format)
        ? filters.formats.filter((selectedFormat) => selectedFormat !== format)
        : [...filters.formats, format],
});
