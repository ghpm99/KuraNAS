import { useEffect, useState } from 'react';

export const useHeader = (showClock: boolean) => {
    const [mobileOpen, setMobileOpen] = useState(false);
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
        mobileOpen,
        closeMobileMenu: () => setMobileOpen(false),
        openMobileMenu: () => setMobileOpen(true),
    };
};
