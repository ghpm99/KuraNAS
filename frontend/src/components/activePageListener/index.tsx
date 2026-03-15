import { appRoutes, isImageRoute, isMusicRoute, isVideoRoute } from '@/app/routes';
import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { pages, useUI } from '../providers/uiProvider/uiContext';

const getPageFromPath = (pathname: string): pages => {
	if (isMusicRoute(pathname)) {
		return 'music';
	}

	if (isVideoRoute(pathname)) {
		return 'videos';
	}

	if (isImageRoute(pathname)) {
		return 'images';
	}

	switch (pathname) {
		case appRoutes.home:
			return 'home';
		case appRoutes.files:
			return 'files';
		case appRoutes.favorites:
		case appRoutes.legacyFavorites:
			return 'favorites';
		case appRoutes.settings:
			return 'settings';
		case appRoutes.activityDiary:
			return 'activity';
		case appRoutes.analytics:
			return 'analytics';
		case appRoutes.about:
			return 'about';
		default:
			return 'unknown';
	}
};

const ActivePageListener = () => {
	const location = useLocation();

	const { setActivePage } = useUI();

	useEffect(() => {
		setActivePage(getPageFromPath(location.pathname));
	}, [location.pathname, setActivePage]);

	return null;
};

export default ActivePageListener;
