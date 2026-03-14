import { getMusicRoute } from '@/app/routes';
import { getMusicSectionFromPath, musicNavigationItems } from '@/components/music/navigation';
import { useLocation } from 'react-router-dom';

export const useMusicNavigation = () => {
	const location = useLocation();
	const currentSection = getMusicSectionFromPath(location.pathname);

	return {
		currentSection,
		items: musicNavigationItems.map((item) => ({
			...item,
			href: getMusicRoute(item.key),
			isActive: item.key === currentSection,
		})),
	};
};
