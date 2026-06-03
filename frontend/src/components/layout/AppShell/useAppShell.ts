import { useGlobalMusic } from '@/features/music/providers/GlobalMusicProvider';

export const useAppShell = () => {
    const { hasQueue } = useGlobalMusic();

    return {
        hasQueue,
        showClock: false,
    };
};
