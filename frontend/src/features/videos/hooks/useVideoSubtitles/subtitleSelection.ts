import type { VideoSubtitleDto } from '@/service/videoPlayback';

export const getPrimaryLanguage = (languageTag: string): string =>
    (languageTag.split(/[-_]/)[0] ?? '').toLowerCase();

const isSameLanguage = (subtitle: VideoSubtitleDto, languageTag: string): boolean => {
    const primaryLanguage = getPrimaryLanguage(languageTag);
    return primaryLanguage !== '' && getPrimaryLanguage(subtitle.language) === primaryLanguage;
};

const findPreferredInLanguage = (
    subtitles: VideoSubtitleDto[],
    languageTag: string
): VideoSubtitleDto | undefined => {
    const subtitlesInLanguage = subtitles.filter((subtitle) =>
        isSameLanguage(subtitle, languageTag)
    );
    return subtitlesInLanguage.find((subtitle) => !subtitle.is_forced) ?? subtitlesInLanguage[0];
};

export const chooseInitialSubtitleId = (
    subtitles: VideoSubtitleDto[],
    rememberedLanguage: string | null,
    interfaceLanguage: string
): number | null => {
    if (rememberedLanguage) {
        const rememberedSubtitle = findPreferredInLanguage(subtitles, rememberedLanguage);
        if (rememberedSubtitle) {
            return rememberedSubtitle.id;
        }
    }
    const forcedInterfaceSubtitle = subtitles.find(
        (subtitle) => subtitle.is_forced && isSameLanguage(subtitle, interfaceLanguage)
    );
    return forcedInterfaceSubtitle?.id ?? null;
};
