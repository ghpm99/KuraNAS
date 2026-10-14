import { useCallback, useEffect, useRef, useState } from 'react';
import { loadAutoplayNextPreference, saveAutoplayNextPreference } from './autoplayNextPreference';

export const UP_NEXT_COUNTDOWN_SECONDS = 8;

const ONE_SECOND_MS = 1000;

type UseUpNextAutoplayParams = {
    currentVideoId: number | null;
    hasNextVideo: boolean;
    isPlaying: boolean;
    defaultAutoplayEnabled: boolean;
    onPlayNext: () => void | Promise<void>;
};

export type UpNextAutoplay = {
    isOpen: boolean;
    isCounting: boolean;
    secondsRemaining: number;
    isAutoplayEnabled: boolean;
    open: () => void;
    cancel: () => void;
    playNow: () => void;
    setAutoplayEnabled: (isEnabled: boolean) => void;
};

const useUpNextAutoplay = ({
    currentVideoId,
    hasNextVideo,
    isPlaying,
    defaultAutoplayEnabled,
    onPlayNext,
}: UseUpNextAutoplayParams): UpNextAutoplay => {
    const [openedForVideo, setOpenedForVideo] = useState<{ videoId: number | null } | null>(null);
    const [secondsRemaining, setSecondsRemaining] = useState(UP_NEXT_COUNTDOWN_SECONDS);
    const [isAutoplayEnabled, setIsAutoplayEnabled] = useState(() =>
        loadAutoplayNextPreference(defaultAutoplayEnabled)
    );
    const onPlayNextRef = useRef(onPlayNext);

    if (isPlaying && openedForVideo !== null) {
        setOpenedForVideo(null);
    }

    const isOpen =
        hasNextVideo && openedForVideo !== null && openedForVideo.videoId === currentVideoId;
    const isCounting = isOpen && isAutoplayEnabled;

    useEffect(() => {
        onPlayNextRef.current = onPlayNext;
    }, [onPlayNext]);

    const cancel = useCallback(() => setOpenedForVideo(null), []);

    const open = useCallback(() => {
        setSecondsRemaining(UP_NEXT_COUNTDOWN_SECONDS);
        setOpenedForVideo({ videoId: currentVideoId });
    }, [currentVideoId]);

    const playNow = useCallback(() => {
        setOpenedForVideo(null);
        void onPlayNextRef.current();
    }, []);

    const setAutoplayEnabled = useCallback((isEnabled: boolean) => {
        setIsAutoplayEnabled(isEnabled);
        setSecondsRemaining(UP_NEXT_COUNTDOWN_SECONDS);
        saveAutoplayNextPreference(isEnabled);
    }, []);

    useEffect(() => {
        if (!isCounting) return;
        const tickTimerId = window.setTimeout(() => {
            if (secondsRemaining <= 1) {
                playNow();
                return;
            }
            setSecondsRemaining(secondsRemaining - 1);
        }, ONE_SECOND_MS);
        return () => window.clearTimeout(tickTimerId);
    }, [isCounting, playNow, secondsRemaining]);

    return {
        isOpen,
        isCounting,
        secondsRemaining,
        isAutoplayEnabled,
        open,
        cancel,
        playNow,
        setAutoplayEnabled,
    };
};

export default useUpNextAutoplay;
