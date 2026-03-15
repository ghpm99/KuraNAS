import { getImageRoute } from '@/app/routes';
import { getImageSectionFromPath, imageNavigationItems } from '@/components/images/navigation';
import { useLocation } from 'react-router-dom';

export const useImageNavigation = () => {
	const location = useLocation();
	const currentSection = getImageSectionFromPath(location.pathname);

	return {
		currentSection,
		items: imageNavigationItems.map((item) => ({
			...item,
			href: getImageRoute(item.key),
			isActive: item.key === currentSection,
		})),
	};
};
