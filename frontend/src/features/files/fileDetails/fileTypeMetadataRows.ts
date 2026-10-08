import type { AudioSummary, ImageSummary, VideoSummary } from '@/types/fileTypeMetadata';

type MetadataRow = {
    labelKey: string;
    value: string;
};

type MetadataEntry = [labelKey: string, value: string | null];

const bitsPerSecondThreshold = 10_000;
const secondsPerMinute = 60;
const minutesPerHour = 60;

const isPositiveNumber = (value: number | undefined): value is number =>
    typeof value === 'number' && Number.isFinite(value) && value > 0;

const textOrNull = (value: string | undefined): string | null => {
    const trimmedValue = (value ?? '').trim();
    return trimmedValue === '' ? null : trimmedValue;
};

const joinPresent = (parts: Array<string | null>, separator: string): string | null => {
    const presentParts = parts.filter((part): part is string => part !== null);
    return presentParts.length === 0 ? null : presentParts.join(separator);
};

const formatDimensions = (width: number | undefined, height: number | undefined): string | null =>
    isPositiveNumber(width) && isPositiveNumber(height) ? `${width} × ${height}` : null;

const formatExposureTime = (exposureTime: number | undefined): string | null => {
    if (!isPositiveNumber(exposureTime)) return null;
    if (exposureTime >= 1) return `${exposureTime} s`;
    return `1/${Math.round(1 / exposureTime)} s`;
};

const formatAperture = (fNumber: number | undefined): string | null =>
    isPositiveNumber(fNumber) ? `f/${fNumber}` : null;

const formatFocalLength = (focalLength: number | undefined): string | null =>
    isPositiveNumber(focalLength) ? `${focalLength} mm` : null;

const formatNumber = (value: number | undefined): string | null =>
    isPositiveNumber(value) ? String(value) : null;

export const formatClock = (totalSeconds: number): string => {
    const wholeSeconds = Math.round(totalSeconds);
    const seconds = wholeSeconds % secondsPerMinute;
    const totalMinutes = Math.floor(wholeSeconds / secondsPerMinute);
    const minutes = totalMinutes % minutesPerHour;
    const hours = Math.floor(totalMinutes / minutesPerHour);
    const paddedSeconds = String(seconds).padStart(2, '0');
    if (hours === 0) return `${minutes}:${paddedSeconds}`;
    return `${hours}:${String(minutes).padStart(2, '0')}:${paddedSeconds}`;
};

const formatDurationSeconds = (seconds: number | undefined): string | null =>
    isPositiveNumber(seconds) ? formatClock(seconds) : null;

const formatBitrate = (rawBitrate: number | undefined): string | null => {
    if (!isPositiveNumber(rawBitrate)) return null;
    const kilobitsPerSecond = rawBitrate >= bitsPerSecondThreshold ? rawBitrate / 1000 : rawBitrate;
    return `${Math.round(kilobitsPerSecond)} kbps`;
};

const formatFrameRate = (frameRate: number | undefined): string | null =>
    isPositiveNumber(frameRate) ? `${Math.round(frameRate * 100) / 100} fps` : null;

const formatHertz = (hertz: number | undefined): string | null =>
    isPositiveNumber(hertz) ? `${hertz} Hz` : null;

const toRows = (entries: MetadataEntry[]): MetadataRow[] =>
    entries.flatMap(([labelKey, value]) => (value === null ? [] : [{ labelKey, value }]));

export const buildImageRows = (summary: ImageSummary): MetadataRow[] =>
    toRows([
        ['FILE_METADATA_DIMENSIONS', formatDimensions(summary.width, summary.height)],
        [
            'FILE_METADATA_CAMERA',
            joinPresent([textOrNull(summary.make), textOrNull(summary.model)], ' '),
        ],
        ['FILE_METADATA_LENS', textOrNull(summary.lens_model)],
        ['FILE_METADATA_TAKEN_AT', textOrNull(summary.datetime_original)],
        ['FILE_METADATA_EXPOSURE', formatExposureTime(summary.exposure_time)],
        ['FILE_METADATA_APERTURE', formatAperture(summary.f_number)],
        ['FILE_METADATA_ISO', formatNumber(summary.iso)],
        ['FILE_METADATA_FOCAL_LENGTH', formatFocalLength(summary.focal_length)],
    ]);

export const buildAudioRows = (summary: AudioSummary): MetadataRow[] =>
    toRows([
        ['FILE_METADATA_TITLE_TAG', textOrNull(summary.title)],
        ['FILE_METADATA_ARTIST', textOrNull(summary.artist)],
        ['FILE_METADATA_ALBUM', textOrNull(summary.album)],
        ['FILE_METADATA_GENRE', textOrNull(summary.genre)],
        ['FILE_METADATA_YEAR', textOrNull(summary.year)],
        ['FILE_METADATA_TRACK', textOrNull(summary.track_number)],
        ['FILE_METADATA_DURATION', formatDurationSeconds(summary.length)],
        ['FILE_METADATA_BITRATE', formatBitrate(summary.bitrate)],
        ['FILE_METADATA_SAMPLE_RATE', formatHertz(summary.sample_rate)],
        ['FILE_METADATA_CHANNELS', formatNumber(summary.channels)],
    ]);

const parseNumericText = (text: string | undefined): number | undefined => {
    const parsedNumber = Number.parseFloat(text ?? '');
    return Number.isNaN(parsedNumber) ? undefined : parsedNumber;
};

export const buildVideoRows = (summary: VideoSummary): MetadataRow[] =>
    toRows([
        ['FILE_METADATA_DURATION', formatDurationSeconds(parseNumericText(summary.duration))],
        ['FILE_METADATA_DIMENSIONS', formatDimensions(summary.width, summary.height)],
        ['FILE_METADATA_VIDEO_CODEC', textOrNull(summary.codec_name)],
        ['FILE_METADATA_FRAME_RATE', formatFrameRate(summary.frame_rate)],
        ['FILE_METADATA_BITRATE', formatBitrate(parseNumericText(summary.bit_rate))],
        ['FILE_METADATA_AUDIO_CODEC', textOrNull(summary.audio_codec)],
        ['FILE_METADATA_CONTAINER', textOrNull(summary.format_name)],
    ]);
