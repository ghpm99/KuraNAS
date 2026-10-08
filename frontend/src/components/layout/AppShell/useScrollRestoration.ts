import { useEffect, useLayoutEffect, useRef, type RefObject } from 'react';
import { useLocation, useNavigationType } from 'react-router-dom';

const MAX_RESTORE_FRAMES = 30;

const savedScrollTopByLocationKey = new Map<string, number>();

export const clearSavedScrollPositions = () => savedScrollTopByLocationKey.clear();

const restoreScrollTop = (scrollElement: HTMLElement, targetScrollTop: number) => {
    let remainingFrames = MAX_RESTORE_FRAMES;
    let frameId = 0;

    const attemptRestore = () => {
        scrollElement.scrollTop = targetScrollTop;
        const maxScrollTop = scrollElement.scrollHeight - scrollElement.clientHeight;
        const isTargetReachable = maxScrollTop >= targetScrollTop;
        remainingFrames -= 1;
        if (isTargetReachable || remainingFrames <= 0) return;
        frameId = requestAnimationFrame(attemptRestore);
    };

    attemptRestore();
    return () => cancelAnimationFrame(frameId);
};

export const useScrollRestoration = (scrollElementRef: RefObject<HTMLElement | null>) => {
    const location = useLocation();
    const navigationType = useNavigationType();
    const currentLocationKeyRef = useRef(location.key);
    const previousPathnameRef = useRef(location.pathname);

    useLayoutEffect(() => {
        currentLocationKeyRef.current = location.key;
    }, [location.key]);

    useEffect(() => {
        const scrollElement = scrollElementRef.current;
        if (!scrollElement) return;

        const rememberScrollTop = () =>
            savedScrollTopByLocationKey.set(currentLocationKeyRef.current, scrollElement.scrollTop);

        scrollElement.addEventListener('scroll', rememberScrollTop, { passive: true });
        return () => scrollElement.removeEventListener('scroll', rememberScrollTop);
    }, [scrollElementRef]);

    useEffect(() => {
        const scrollElement = scrollElementRef.current;
        const hasPathnameChanged = previousPathnameRef.current !== location.pathname;
        previousPathnameRef.current = location.pathname;
        if (!scrollElement) return;

        if (navigationType === 'POP') {
            const savedScrollTop = savedScrollTopByLocationKey.get(location.key) ?? 0;
            return restoreScrollTop(scrollElement, savedScrollTop);
        }

        const shouldResetScroll = navigationType === 'PUSH' || hasPathnameChanged;
        if (shouldResetScroll) scrollElement.scrollTop = 0;
    }, [location.key, location.pathname, navigationType, scrollElementRef]);
};
