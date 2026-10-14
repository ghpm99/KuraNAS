import { useEffect, useState } from 'react';

const CONTROLS_IDLE_TIMEOUT_MS = 3000;
const USER_ACTIVITY_EVENTS = ['mousemove', 'keydown', 'touchstart'] as const;

const useControlsAutoHide = (isPlaying: boolean) => {
    const [isIdle, setIsIdle] = useState(false);

    useEffect(() => {
        if (!isPlaying) return;
        const markAsIdle = () => setIsIdle(true);
        let idleTimeoutId = setTimeout(markAsIdle, CONTROLS_IDLE_TIMEOUT_MS);
        const handleUserActivity = () => {
            setIsIdle(false);
            clearTimeout(idleTimeoutId);
            idleTimeoutId = setTimeout(markAsIdle, CONTROLS_IDLE_TIMEOUT_MS);
        };

        USER_ACTIVITY_EVENTS.forEach((eventName) =>
            document.addEventListener(eventName, handleUserActivity)
        );
        return () => {
            USER_ACTIVITY_EVENTS.forEach((eventName) =>
                document.removeEventListener(eventName, handleUserActivity)
            );
            clearTimeout(idleTimeoutId);
            setIsIdle(false);
        };
    }, [isPlaying]);

    return { areControlsVisible: !isIdle };
};

export default useControlsAutoHide;
