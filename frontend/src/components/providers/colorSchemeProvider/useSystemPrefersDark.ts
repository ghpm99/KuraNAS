import { useEffect, useState } from 'react';
import { readSystemPrefersDark, systemDarkMediaQuery } from './colorScheme';

export const useSystemPrefersDark = (): boolean => {
    const [systemPrefersDark, setSystemPrefersDark] = useState(readSystemPrefersDark);

    useEffect(() => {
        if (typeof window.matchMedia !== 'function') {
            return;
        }
        const mediaQueryList = window.matchMedia(systemDarkMediaQuery);
        const handleChange = (event: MediaQueryListEvent) => setSystemPrefersDark(event.matches);
        if (typeof mediaQueryList.addEventListener === 'function') {
            mediaQueryList.addEventListener('change', handleChange);
            return () => mediaQueryList.removeEventListener('change', handleChange);
        }
        mediaQueryList.addListener(handleChange);
        return () => mediaQueryList.removeListener(handleChange);
    }, []);

    return systemPrefersDark;
};
