import { useGlobalMusic } from '@/components/providers/GlobalMusicProvider';
import { useUI } from '@/components/providers/uiProvider/uiContext';

export const useAppShell = () => {
	const { activePage } = useUI();
	const { hasQueue } = useGlobalMusic();

	return {
		hasQueue,
		showClock: activePage === 'activity',
	};
};
