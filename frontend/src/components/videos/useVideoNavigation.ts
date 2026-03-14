import { getVideoRoute } from '@/app/routes';
import { getVideoSectionFromPath, videoNavigationItems } from '@/components/videos/navigation';
import { useLocation } from 'react-router-dom';

export const useVideoNavigation = () => {
	const location = useLocation();
	const currentSection = getVideoSectionFromPath(location.pathname);

	return {
		currentSection,
		items: videoNavigationItems.map((item) => ({
			...item,
			href: getVideoRoute(item.key),
			isActive: item.key === currentSection,
		})),
	};
};
