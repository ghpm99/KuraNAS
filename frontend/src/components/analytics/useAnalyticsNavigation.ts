import { getAnalyticsRoute } from '@/app/routes';
import {
    analyticsNavigationItems,
    getAnalyticsSectionFromPath,
} from '@/components/analytics/navigation';
import { useLocation } from 'react-router-dom';

export const useAnalyticsNavigation = () => {
    const location = useLocation();
    const currentSection = getAnalyticsSectionFromPath(location.pathname);

    return {
        currentSection,
        items: analyticsNavigationItems.map((item) => ({
            ...item,
            href: getAnalyticsRoute(item.key),
            isActive: item.key === currentSection,
        })),
    };
};
