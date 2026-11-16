import { useMutation } from '@tanstack/react-query';
import { useCallback, useEffect, useRef, useState } from 'react';
import { sendVacuumRemoteControl } from '@/service/vacuum';
import type { VacuumDirection } from '@/types/vacuum';
import { useVacuumFeedback } from './useVacuumFeedback';
import type { VacuumMovementDirection } from './vacuumRemoteControlKeys';

export const useVacuumRemoteControl = () => {
    const { notifyFailure } = useVacuumFeedback();
    const movingDirectionRef = useRef<VacuumMovementDirection | null>(null);
    const [movingDirection, setMovingDirection] = useState<VacuumMovementDirection | null>(null);

    const { mutate: sendDirection } = useMutation({
        mutationFn: (direction: VacuumDirection) => sendVacuumRemoteControl(direction),
        onError: notifyFailure,
    });

    const startMoving = useCallback(
        (direction: VacuumMovementDirection) => {
            if (movingDirectionRef.current === direction) return;
            movingDirectionRef.current = direction;
            setMovingDirection(direction);
            sendDirection(direction);
        },
        [sendDirection]
    );

    const stopMoving = useCallback(
        (direction?: VacuumMovementDirection) => {
            const currentDirection = movingDirectionRef.current;
            if (currentDirection === null) return;
            if (direction !== undefined && direction !== currentDirection) return;
            movingDirectionRef.current = null;
            setMovingDirection(null);
            sendDirection('stop');
        },
        [sendDirection]
    );

    useEffect(() => {
        const stopWhenTabIsHidden = () => {
            if (document.hidden) stopMoving();
        };
        const stopWhenWindowLosesFocus = () => stopMoving();

        document.addEventListener('visibilitychange', stopWhenTabIsHidden);
        window.addEventListener('blur', stopWhenWindowLosesFocus);
        return () => {
            document.removeEventListener('visibilitychange', stopWhenTabIsHidden);
            window.removeEventListener('blur', stopWhenWindowLosesFocus);
            stopMoving();
        };
    }, [stopMoving]);

    return { movingDirection, startMoving, stopMoving };
};
