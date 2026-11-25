import { useLayoutEffect, useState, type RefObject } from 'react';

export const useMeasuredElementWidth = (elementRef: RefObject<HTMLElement | null>): number => {
    const [widthCssPixels, setWidthCssPixels] = useState(0);

    useLayoutEffect(() => {
        const element = elementRef.current;
        if (element === null) return;
        const measure = () => setWidthCssPixels(Math.round(element.getBoundingClientRect().width));
        measure();
        if (typeof ResizeObserver === 'undefined') {
            window.addEventListener('resize', measure);
            return () => window.removeEventListener('resize', measure);
        }
        const observer = new ResizeObserver(measure);
        observer.observe(element);
        return () => observer.disconnect();
    }, [elementRef]);

    return widthCssPixels;
};
