import type { VideoSubtitleDto } from '@/service/videoPlayback';

export const buildSubtitleLabel = (
    subtitle: VideoSubtitleDto,
    translate: (key: string) => string
): string => {
    const name = subtitle.language || subtitle.label;
    const formatSuffix = subtitle.format.toUpperCase();
    const forcedSuffix = subtitle.is_forced ? ` · ${translate('VIDEO_SUBTITLES_FORCED')}` : '';
    return `${name} (${formatSuffix})${forcedSuffix}`;
};
