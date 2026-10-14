import { useQuery } from '@tanstack/react-query';
import { useCallback, useEffect, useMemo, useRef, useState, type RefObject } from 'react';
import {
    getVideoSubtitleKey,
    getVideoSubtitles,
    type VideoSubtitleDto,
} from '@/service/videoPlayback';
import { chooseInitialSubtitleKey, findRememberedLanguageSubtitleKey } from './subtitleSelection';
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
    subtitleKey: string | null;
}

const noSubtitles: VideoSubtitleDto[] = [];

const applySubtitleModes = (video: HTMLVideoElement, selectedSubtitleKey: string | null) => {
    video.querySelectorAll('track').forEach((trackElement) => {
        if (!trackElement.track) {
            return;
        }
        const isSelected = trackElement.dataset.subtitleKey === selectedSubtitleKey;
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
    const lastEnabledSubtitleKeyRef = useRef<string | null>(null);

    const initialSubtitleKey = useMemo(
        () =>
            chooseInitialSubtitleKey(subtitles, loadVideoSubtitlePreferences(), interfaceLanguage),
        [subtitles, interfaceLanguage]
    );
    const selectedSubtitleKey =
        userChoice?.videoId === videoId ? userChoice.subtitleKey : initialSubtitleKey;

    useEffect(() => {
        const video = videoRef.current;
        if (video) {
            applySubtitleModes(video, selectedSubtitleKey);
        }
    }, [subtitles, selectedSubtitleKey, videoRef]);

    const selectSubtitle = useCallback(
        (subtitleKey: string | null) => {
            lastEnabledSubtitleKeyRef.current = subtitleKey ?? selectedSubtitleKey;
            setUserChoice({ videoId, subtitleKey });
            if (subtitleKey === null) {
                const selectedSubtitle = subtitles.find(
                    (subtitle) => getVideoSubtitleKey(subtitle) === selectedSubtitleKey
                );
                const lastLanguage =
                    selectedSubtitle?.language || loadVideoSubtitlePreferences()?.language || '';
                saveVideoSubtitlePreferences({ isEnabled: false, language: lastLanguage });
                return;
            }
            const chosenSubtitle = subtitles.find(
                (subtitle) => getVideoSubtitleKey(subtitle) === subtitleKey
            );
            if (chosenSubtitle?.language) {
                saveVideoSubtitlePreferences({
                    isEnabled: true,
                    language: chosenSubtitle.language,
                });
            }
        },
        [selectedSubtitleKey, subtitles, videoId]
    );

    const toggleSubtitles = useCallback((): SubtitleToggleOutcome => {
        if (subtitles.length === 0) {
            return 'unavailable';
        }
        if (selectedSubtitleKey !== null) {
            selectSubtitle(null);
            return 'disabled';
        }
        const lastEnabledSubtitle = subtitles.find(
            (subtitle) => getVideoSubtitleKey(subtitle) === lastEnabledSubtitleKeyRef.current
        );
        const rememberedSubtitleKey = findRememberedLanguageSubtitleKey(
            subtitles,
            loadVideoSubtitlePreferences()
        );
        const firstSubtitle = subtitles[0];
        selectSubtitle(
            (lastEnabledSubtitle && getVideoSubtitleKey(lastEnabledSubtitle)) ??
                rememberedSubtitleKey ??
                (firstSubtitle && getVideoSubtitleKey(firstSubtitle)) ??
                null
        );
        return 'enabled';
    }, [selectSubtitle, selectedSubtitleKey, subtitles]);

    return { subtitles, selectedSubtitleKey, selectSubtitle, toggleSubtitles };
}
