import { useCallback, useEffect, useRef, type PointerEvent } from 'react';

const longPressDelayInMilliseconds = 500;

export const useLongPress = <TTarget>(onLongPress: (target: TTarget) => void) => {
    const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const hasFiredRef = useRef(false);

    const cancelLongPress = useCallback(() => {
        if (timerRef.current === null) {
            return;
        }
        clearTimeout(timerRef.current);
        timerRef.current = null;
    }, []);

    const startLongPress = useCallback(
        (event: PointerEvent, target: TTarget) => {
            if (event.pointerType !== 'touch') {
                return;
            }
            hasFiredRef.current = false;
            cancelLongPress();
            timerRef.current = setTimeout(() => {
                timerRef.current = null;
                hasFiredRef.current = true;
                onLongPress(target);
            }, longPressDelayInMilliseconds);
        },
        [cancelLongPress, onLongPress]
    );

    const consumeFiredLongPress = useCallback(() => {
        const hasFired = hasFiredRef.current;
        hasFiredRef.current = false;
        return hasFired;
    }, []);

    useEffect(() => cancelLongPress, [cancelLongPress]);

    return { startLongPress, cancelLongPress, consumeFiredLongPress };
};
