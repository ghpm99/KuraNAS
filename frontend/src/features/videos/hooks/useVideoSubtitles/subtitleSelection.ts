import type { VideoSubtitleDto } from '@/service/videoPlayback';
import type { VideoSubtitlePreferences } from './videoSubtitlePreferences';

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

export const findRememberedLanguageSubtitleId = (
    subtitles: VideoSubtitleDto[],
    preferences: VideoSubtitlePreferences | null
): number | null => {
    if (!preferences?.language) {
        return null;
    }
    return findPreferredInLanguage(subtitles, preferences.language)?.id ?? null;
};

export const chooseInitialSubtitleId = (
    subtitles: VideoSubtitleDto[],
    preferences: VideoSubtitlePreferences | null,
    interfaceLanguage: string
): number | null => {
    if (preferences?.isEnabled) {
        const rememberedSubtitleId = findRememberedLanguageSubtitleId(subtitles, preferences);
        if (rememberedSubtitleId !== null) {
            return rememberedSubtitleId;
        }
    }
    const forcedInterfaceSubtitle = subtitles.find(
        (subtitle) => subtitle.is_forced && isSameLanguage(subtitle, interfaceLanguage)
    );
    return forcedInterfaceSubtitle?.id ?? null;
};
