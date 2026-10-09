import { useEffect, useRef, type RefObject } from 'react';
import { useLocation } from 'react-router-dom';

const pageHeadingSelector = 'h1';

export const useRouteFocus = (
    mainRef: RefObject<HTMLElement | null>,
    announcerRef: RefObject<HTMLElement | null>
) => {
    const { pathname } = useLocation();
    const previousPathnameRef = useRef(pathname);

    useEffect(() => {
        if (previousPathnameRef.current === pathname) return;
        previousPathnameRef.current = pathname;

        const mainElement = mainRef.current;
        if (!mainElement) return;
        const pageHeading = mainElement.querySelector<HTMLElement>(pageHeadingSelector);
        const focusTarget = pageHeading ?? mainElement;
        focusTarget.focus({ preventScroll: true });
        if (announcerRef.current) {
            announcerRef.current.textContent = pageHeading?.textContent || document.title;
        }
    }, [pathname, mainRef, announcerRef]);
};
