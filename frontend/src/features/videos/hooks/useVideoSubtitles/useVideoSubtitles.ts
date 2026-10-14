import { useQuery } from '@tanstack/react-query';
import { useCallback, useEffect, useMemo, useRef, useState, type RefObject } from 'react';
import { getVideoSubtitles, type VideoSubtitleDto } from '@/service/videoPlayback';
import { chooseInitialSubtitleId, findRememberedLanguageSubtitleId } from './subtitleSelection';
import {
    loadVideoSubtitlePreferences,
    saveVideoSubtitlePreferences,
} from './videoSubtitlePreferences';

export type SubtitleToggleOutcome = 'enabled' | 'disabled' | 'unavailable';

interface UseVideoSubtitlesOptions {
    videoId: number | null;
    videoRef: RefObject<HTMLVideoElement | null>;
    interfaceLanguage: string;
}

interface UserSubtitleChoice {
    videoId: number | null;
    subtitleId: number | null;
}

const noSubtitles: VideoSubtitleDto[] = [];

const applySubtitleModes = (video: HTMLVideoElement, selectedSubtitleId: number | null) => {
    video.querySelectorAll('track').forEach((trackElement) => {
        if (!trackElement.track) {
            return;
        }
        const isSelected = Number(trackElement.dataset.subtitleId) === selectedSubtitleId;
        trackElement.track.mode = isSelected ? 'showing' : 'disabled';
    });
};

export default function useVideoSubtitles({
    videoId,
    videoRef,
    interfaceLanguage,
}: UseVideoSubtitlesOptions) {
    const { data: fetchedSubtitles } = useQuery({
        queryKey: ['video-subtitles', videoId],
        queryFn: () => getVideoSubtitles(videoId as number),
        enabled: videoId !== null,
        retry: false,
        refetchOnWindowFocus: false,
    });
    const subtitles = Array.isArray(fetchedSubtitles) ? fetchedSubtitles : noSubtitles;
    const [userChoice, setUserChoice] = useState<UserSubtitleChoice | null>(null);
    const lastEnabledSubtitleIdRef = useRef<number | null>(null);

    const initialSubtitleId = useMemo(
        () => chooseInitialSubtitleId(subtitles, loadVideoSubtitlePreferences(), interfaceLanguage),
        [subtitles, interfaceLanguage]
    );
    const selectedSubtitleId =
        userChoice?.videoId === videoId ? userChoice.subtitleId : initialSubtitleId;

    useEffect(() => {
        const video = videoRef.current;
        if (video) {
            applySubtitleModes(video, selectedSubtitleId);
        }
    }, [subtitles, selectedSubtitleId, videoRef]);

    const selectSubtitle = useCallback(
        (subtitleId: number | null) => {
            lastEnabledSubtitleIdRef.current = subtitleId ?? selectedSubtitleId;
            setUserChoice({ videoId, subtitleId });
            if (subtitleId === null) {
                const selectedSubtitle = subtitles.find(
                    (subtitle) => subtitle.id === selectedSubtitleId
                );
                const lastLanguage =
                    selectedSubtitle?.language || loadVideoSubtitlePreferences()?.language || '';
                saveVideoSubtitlePreferences({ isEnabled: false, language: lastLanguage });
                return;
            }
            const chosenSubtitle = subtitles.find((subtitle) => subtitle.id === subtitleId);
            if (chosenSubtitle?.language) {
                saveVideoSubtitlePreferences({
                    isEnabled: true,
                    language: chosenSubtitle.language,
                });
            }
        },
        [selectedSubtitleId, subtitles, videoId]
    );

    const toggleSubtitles = useCallback((): SubtitleToggleOutcome => {
        if (subtitles.length === 0) {
            return 'unavailable';
        }
        if (selectedSubtitleId !== null) {
            selectSubtitle(null);
            return 'disabled';
        }
        const lastEnabledSubtitle = subtitles.find(
            (subtitle) => subtitle.id === lastEnabledSubtitleIdRef.current
        );
        const rememberedSubtitleId = findRememberedLanguageSubtitleId(
            subtitles,
            loadVideoSubtitlePreferences()
        );
        selectSubtitle(lastEnabledSubtitle?.id ?? rememberedSubtitleId ?? subtitles[0]?.id ?? null);
        return 'enabled';
    }, [selectSubtitle, selectedSubtitleId, subtitles]);

    return { subtitles, selectedSubtitleId, selectSubtitle, toggleSubtitles };
}
