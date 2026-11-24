import type { VideoAudioTrackDto } from '@/service/videoPlayback';
import { getPrimaryLanguage } from '@/features/videos/hooks/useVideoSubtitles/subtitleSelection';

export const findDefaultAudioTrack = (
    audioTracks: VideoAudioTrackDto[]
): VideoAudioTrackDto | undefined =>
    audioTracks.find((audioTrack) => audioTrack.is_default) ?? audioTracks[0];

export const findAudioStreamIndexToStartWith = (
    audioTracks: VideoAudioTrackDto[],
    rememberedLanguage: string | null
): number | null => {
    if (audioTracks.length < 2 || !rememberedLanguage) return null;
    const rememberedPrimaryLanguage = getPrimaryLanguage(rememberedLanguage);
    const isRememberedLanguage = (audioTrack: VideoAudioTrackDto) =>
        getPrimaryLanguage(audioTrack.language) === rememberedPrimaryLanguage;
    const defaultAudioTrack = findDefaultAudioTrack(audioTracks);
    if (defaultAudioTrack && isRememberedLanguage(defaultAudioTrack)) return null;
    return audioTracks.find(isRememberedLanguage)?.stream_index ?? null;
};
