import { useEffect, useState } from 'react';

export const useHeader = (showClock: boolean) => {
    const [currentTime, setCurrentTime] = useState(() => new Date());

    useEffect(() => {
        if (!showClock) {
            return undefined;
        }

        const intervalId = window.setInterval(() => {
            setCurrentTime(new Date());
        }, 1000);

        return () => {
            window.clearInterval(intervalId);
        };
    }, [showClock]);

    return {
        currentTime,
    };
};
