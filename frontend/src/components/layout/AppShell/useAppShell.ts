import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';

export const useAppShell = () => {
    const { hasQueue } = useGlobalMusic();

    return {
        hasQueue,
        showClock: false,
    };
};
