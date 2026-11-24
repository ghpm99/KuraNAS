import type { VideoAudioTrackDto } from '@/service/videoPlayback';

const translateLanguageCode = (languageCode: string, interfaceLanguage: string): string => {
    try {
        const displayNames = new Intl.DisplayNames([interfaceLanguage || 'en'], {
            type: 'language',
        });
        return displayNames.of(languageCode) ?? languageCode;
    } catch {
        return languageCode;
    }
};

export const buildAudioTrackLabel = (
    audioTrack: VideoAudioTrackDto,
    interfaceLanguage: string,
    translate: (key: string, options?: Record<string, string>) => string
): string => {
    const languageName = audioTrack.language
        ? translateLanguageCode(audioTrack.language, interfaceLanguage)
        : translate('VIDEO_AUDIO_TRACK_UNKNOWN_LANGUAGE');
    return audioTrack.title ? `${languageName} · ${audioTrack.title}` : languageName;
};
