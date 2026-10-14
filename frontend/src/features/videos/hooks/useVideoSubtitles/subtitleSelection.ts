import { getVideoSubtitleKey, type VideoSubtitleDto } from '@/service/videoPlayback';
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

export const findRememberedLanguageSubtitleKey = (
    subtitles: VideoSubtitleDto[],
    preferences: VideoSubtitlePreferences | null
): string | null => {
    if (!preferences?.language) {
        return null;
    }
    const preferredSubtitle = findPreferredInLanguage(subtitles, preferences.language);
    return preferredSubtitle ? getVideoSubtitleKey(preferredSubtitle) : null;
};

export const chooseInitialSubtitleKey = (
    subtitles: VideoSubtitleDto[],
    preferences: VideoSubtitlePreferences | null,
    interfaceLanguage: string
): string | null => {
    if (preferences?.isEnabled) {
        const rememberedSubtitleKey = findRememberedLanguageSubtitleKey(subtitles, preferences);
        if (rememberedSubtitleKey !== null) {
            return rememberedSubtitleKey;
        }
    }
    const forcedInterfaceSubtitle = subtitles.find(
        (subtitle) => subtitle.is_forced && isSameLanguage(subtitle, interfaceLanguage)
    );
    return forcedInterfaceSubtitle ? getVideoSubtitleKey(forcedInterfaceSubtitle) : null;
};
