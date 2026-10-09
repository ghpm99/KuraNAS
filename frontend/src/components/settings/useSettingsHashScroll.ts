import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';

export const useSettingsHashScroll = () => {
    const { hash } = useLocation();

    useEffect(() => {
        const anchorId = decodeURIComponent(hash.replace(/^#/, ''));
        if (anchorId === '') {
            return;
        }
        document.getElementById(anchorId)?.scrollIntoView?.({ block: 'start' });
    }, [hash]);
};
