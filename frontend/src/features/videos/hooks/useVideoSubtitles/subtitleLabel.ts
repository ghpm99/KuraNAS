import { isEmbeddedSubtitle, type VideoSubtitleDto } from '@/service/videoPlayback';

export const buildSubtitleLabel = (
    subtitle: VideoSubtitleDto,
    translate: (key: string) => string
): string => {
    const name = subtitle.language || subtitle.label;
    const forcedSuffix = subtitle.is_forced ? ` · ${translate('VIDEO_SUBTITLES_FORCED')}` : '';
    if (isEmbeddedSubtitle(subtitle)) {
        return `${name} · ${translate('VIDEO_SUBTITLES_EMBEDDED')}${forcedSuffix}`;
    }
    return `${name} (${subtitle.format.toUpperCase()})${forcedSuffix}`;
};
