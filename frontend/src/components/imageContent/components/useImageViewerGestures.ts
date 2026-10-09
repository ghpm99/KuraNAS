import { useMemo, useRef, type PointerEvent, type MouseEvent } from 'react';

type Point = { x: number; y: number };

type UseImageViewerGesturesParams = {
    zoom: number;
    pan: Point;
    onZoomChange: (zoom: number) => void;
    onPanChange: (panX: number, panY: number) => void;
    onPrevious: () => void;
    onNext: () => void;
};

const swipeMinimumDistanceInPixels = 60;
const swipeHorizontalDominance = 1.5;
const doubleTapWindowInMs = 300;
const doubleTapMaximumDistanceInPixels = 30;
const doubleClickAfterTapGuardInMs = 500;
const zoomedInLevel = 2;

const distanceBetween = (first: Point, second: Point) =>
    Math.hypot(first.x - second.x, first.y - second.y);

const distanceBetweenFirstTwo = (points: Map<number, Point>) => {
    const [firstPoint, secondPoint] = [...points.values()];
    return firstPoint && secondPoint ? distanceBetween(firstPoint, secondPoint) : 0;
};

const clampPan = (panValue: number, frameSize: number, zoom: number) => {
    if (frameSize <= 0) {
        return panValue;
    }
    const limit = (frameSize * zoom) / 2;
    return Math.min(limit, Math.max(-limit, panValue));
};

export const useImageViewerGestures = ({
    zoom,
    pan,
    onZoomChange,
    onPanChange,
    onPrevious,
    onNext,
}: UseImageViewerGesturesParams) => {
    const activePointers = useRef(new Map<number, Point>());
    const dragOrigin = useRef<{ pointer: Point; pan: Point } | null>(null);
    const pinchOrigin = useRef<{ distance: number; zoom: number } | null>(null);
    const hasPinchedInGesture = useRef(false);
    const lastTap = useRef<{ point: Point; time: number } | null>(null);
    const lastZoomToggleAt = useRef(0);

    return useMemo(() => {
        const toggleZoom = () => {
            lastZoomToggleAt.current = Date.now();
            onZoomChange(zoom > 1 ? 1 : zoomedInLevel);
        };

        const registersDoubleTap = (point: Point) => {
            const previousTap = lastTap.current;
            const now = Date.now();
            const isDoubleTap =
                previousTap !== null &&
                now - previousTap.time <= doubleTapWindowInMs &&
                distanceBetween(previousTap.point, point) <= doubleTapMaximumDistanceInPixels;
            lastTap.current = isDoubleTap ? null : { point, time: now };
            return isDoubleTap;
        };

        const onPointerDown = (event: PointerEvent<HTMLElement>) => {
            const point = { x: event.clientX, y: event.clientY };
            event.currentTarget.setPointerCapture?.(event.pointerId);
            activePointers.current.set(event.pointerId, point);

            if (activePointers.current.size >= 2) {
                pinchOrigin.current = {
                    distance: distanceBetweenFirstTwo(activePointers.current),
                    zoom,
                };
                hasPinchedInGesture.current = true;
                dragOrigin.current = null;
                return;
            }

            hasPinchedInGesture.current = false;
            dragOrigin.current = { pointer: point, pan };
            if (event.pointerType !== 'mouse' && registersDoubleTap(point)) {
                toggleZoom();
            }
        };

        const onPointerMove = (event: PointerEvent<HTMLElement>) => {
            if (!activePointers.current.has(event.pointerId)) {
                return;
            }
            const point = { x: event.clientX, y: event.clientY };
            activePointers.current.set(event.pointerId, point);

            const pinch = pinchOrigin.current;
            if (activePointers.current.size >= 2 && pinch && pinch.distance > 0) {
                const currentDistance = distanceBetweenFirstTwo(activePointers.current);
                onZoomChange(pinch.zoom * (currentDistance / pinch.distance));
                return;
            }

            const drag = dragOrigin.current;
            if (!drag || zoom <= 1) {
                return;
            }
            const frame = event.currentTarget.getBoundingClientRect();
            onPanChange(
                clampPan(drag.pan.x + point.x - drag.pointer.x, frame.width, zoom),
                clampPan(drag.pan.y + point.y - drag.pointer.y, frame.height, zoom)
            );
        };

        const finishSwipe = (point: Point) => {
            const drag = dragOrigin.current;
            if (!drag || zoom > 1 || hasPinchedInGesture.current) {
                return;
            }
            const horizontalDistance = point.x - drag.pointer.x;
            const verticalDistance = point.y - drag.pointer.y;
            const isHorizontalSwipe =
                Math.abs(horizontalDistance) >= swipeMinimumDistanceInPixels &&
                Math.abs(horizontalDistance) >
                    Math.abs(verticalDistance) * swipeHorizontalDominance;
            if (!isHorizontalSwipe) {
                return;
            }
            if (horizontalDistance < 0) {
                onNext();
                return;
            }
            onPrevious();
        };

        const releasePointer = (event: PointerEvent<HTMLElement>) => {
            activePointers.current.delete(event.pointerId);
            pinchOrigin.current = null;
            dragOrigin.current = null;
            const remainingPointer = [...activePointers.current.values()][0];
            if (remainingPointer) {
                dragOrigin.current = { pointer: remainingPointer, pan };
            }
        };

        const onPointerUp = (event: PointerEvent<HTMLElement>) => {
            if (activePointers.current.size === 1) {
                finishSwipe({ x: event.clientX, y: event.clientY });
            }
            releasePointer(event);
        };

        const onDoubleClick = (event: MouseEvent<HTMLElement>) => {
            event.preventDefault();
            if (Date.now() - lastZoomToggleAt.current < doubleClickAfterTapGuardInMs) {
                return;
            }
            toggleZoom();
        };

        return {
            onPointerDown,
            onPointerMove,
            onPointerUp,
            onPointerCancel: releasePointer,
            onDoubleClick,
        };
    }, [zoom, pan, onZoomChange, onPanChange, onPrevious, onNext]);
};
