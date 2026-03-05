import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { pages, useUI } from '../providers/uiProvider/uiContext';

const routeToPageMap: Record<string, pages> = {
	'/': 'files',
	'/starred': 'files',
	'/images': 'images',
	'/music': 'music',
	'/videos': 'videos',
	'/activity-diary': 'activity',
	'/analytics': 'analytics',
	'/about': 'about',
};

const ActivePageListener = () => {
	const location = useLocation();

	const { setActivePage } = useUI();

	useEffect(() => {
		const page = routeToPageMap[location.pathname] || 'unknown';
		setActivePage(page);
	}, [location.pathname, setActivePage]);

	return null;
};

export default ActivePageListener;
