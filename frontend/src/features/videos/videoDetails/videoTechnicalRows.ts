import { buildVideoRows } from '@/features/files/fileDetails/fileTypeMetadataRows';
import type { VideoFileDto, VideoLibraryResolution } from '@/service/videoPlayback';
import type { VideoSummary } from '@/types/fileTypeMetadata';
import { classifyVideoResolution } from '../providers/videoContentProvider/videoResolutionClass';

export type VideoTechnicalRow = {
    labelKey: string;
    value: string;
    resolutionClass?: VideoLibraryResolution;
};

const dimensionsLabelKey = 'FILE_METADATA_DIMENSIONS';

const isPositive = (value: number | null | undefined): value is number =>
    typeof value === 'number' && Number.isFinite(value) && value > 0;

const textOrEmpty = (value: string | null | undefined): string => (value ?? '').trim();

const preferText = (summaryText: string | undefined, listingText: string | null | undefined) =>
    textOrEmpty(summaryText) || textOrEmpty(listingText);

const preferPositive = (
    summaryNumber: number | undefined,
    listingNumber: number | null | undefined
) => (isPositive(summaryNumber) ? summaryNumber : isPositive(listingNumber) ? listingNumber : 0);

const mergeWithListing = (video: VideoFileDto, summary: Partial<VideoSummary>): VideoSummary => ({
    duration:
        textOrEmpty(summary.duration) ||
        (isPositive(video.duration_seconds) ? String(video.duration_seconds) : ''),
    width: preferPositive(summary.width, video.width),
    height: preferPositive(summary.height, video.height),
    codec_name: preferText(summary.codec_name, video.video_codec),
    frame_rate: preferPositive(summary.frame_rate, null),
    bit_rate: textOrEmpty(summary.bit_rate),
    audio_codec: textOrEmpty(summary.audio_codec),
    format_name: textOrEmpty(summary.format_name),
});

export const buildVideoTechnicalRows = (
    video: VideoFileDto,
    summary: Partial<VideoSummary> | undefined
): VideoTechnicalRow[] => {
    const mergedSummary = mergeWithListing(video, summary ?? {});
    const resolutionClass = classifyVideoResolution(mergedSummary.height) ?? undefined;

    return buildVideoRows(mergedSummary).map((row) =>
        row.labelKey === dimensionsLabelKey ? { ...row, resolutionClass } : row
    );
};
