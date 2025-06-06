import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { pages, useUI } from '../hooks/UI/uiContext';

const routeToPageMap: Record<string, pages> = {
	'/': 'files',
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
