import { useCallback, useEffect, useRef, useState } from 'react';

const defaultIntentDelayMs = 300;

const matchesMediaQuery = (query: string): boolean =>
    typeof window !== 'undefined' &&
    typeof window.matchMedia === 'function' &&
    window.matchMedia(query).matches;

const isPreviewAllowedByDevice = (): boolean =>
    !matchesMediaQuery('(hover: none)') && !matchesMediaQuery('(prefers-reduced-motion: reduce)');

type HoverPreviewHandlers = {
    onPointerEnter: () => void;
    onPointerLeave: () => void;
    onFocus: () => void;
    onBlur: () => void;
};

type HoverPreview = {
    isPreviewMounted: boolean;
    isPreviewLoaded: boolean;
    previewHandlers: HoverPreviewHandlers;
    handlePreviewLoad: () => void;
    handlePreviewError: () => void;
};

export const useHoverPreview = (intentDelayMs = defaultIntentDelayMs): HoverPreview => {
    const [isPreviewRequested, setIsPreviewRequested] = useState(false);
    const [isPreviewLoaded, setIsPreviewLoaded] = useState(false);
    const [hasPreviewFailed, setHasPreviewFailed] = useState(false);
    const pendingTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    const cancelPendingRequest = useCallback(() => {
        if (pendingTimerRef.current === null) return;
        clearTimeout(pendingTimerRef.current);
        pendingTimerRef.current = null;
    }, []);

    const schedulePreviewRequest = useCallback(() => {
        if (pendingTimerRef.current !== null) return;
        if (!isPreviewAllowedByDevice()) return;
        pendingTimerRef.current = setTimeout(() => {
            pendingTimerRef.current = null;
            setIsPreviewRequested(true);
        }, intentDelayMs);
    }, [intentDelayMs]);

    useEffect(() => cancelPendingRequest, [cancelPendingRequest]);

    const handlePreviewLoad = useCallback(() => setIsPreviewLoaded(true), []);

    const handlePreviewError = useCallback(() => {
        setHasPreviewFailed(true);
        setIsPreviewLoaded(false);
    }, []);

    return {
        isPreviewMounted: isPreviewRequested && !hasPreviewFailed,
        isPreviewLoaded,
        previewHandlers: {
            onPointerEnter: schedulePreviewRequest,
            onPointerLeave: cancelPendingRequest,
            onFocus: schedulePreviewRequest,
            onBlur: cancelPendingRequest,
        },
        handlePreviewLoad,
        handlePreviewError,
    };
};
