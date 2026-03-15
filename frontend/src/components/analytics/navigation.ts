import { getAnalyticsRoute, type AnalyticsSection } from '@/app/routes';

export type AnalyticsNavigationItem = {
	key: AnalyticsSection;
	labelKey: string;
	descriptionKey: string;
};

export const analyticsNavigationItems: AnalyticsNavigationItem[] = [
	{
		key: 'overview',
		labelKey: 'ANALYTICS_SECTION_OVERVIEW',
		descriptionKey: 'ANALYTICS_SECTION_OVERVIEW_DESCRIPTION',
	},
	{
		key: 'library',
		labelKey: 'ANALYTICS_SECTION_LIBRARY',
		descriptionKey: 'ANALYTICS_SECTION_LIBRARY_DESCRIPTION',
	},
];

const analyticsRouteEntries = analyticsNavigationItems.map((item) => [getAnalyticsRoute(item.key), item.key] as const);

export const getAnalyticsSectionFromPath = (pathname: string): AnalyticsSection => {
	if (pathname === getAnalyticsRoute('overview')) {
		return 'overview';
	}

	const matchedEntry = analyticsRouteEntries.find(([route, section]) => {
		if (section === 'overview') {
			return false;
		}

		return pathname === route || pathname.startsWith(`${route}/`);
	});

	return matchedEntry?.[1] ?? 'overview';
};

export const getAnalyticsSectionMeta = (section: AnalyticsSection) => {
	const matchedItem = analyticsNavigationItems.find((item) => item.key === section);

	if (matchedItem) {
		return matchedItem;
	}

	return analyticsNavigationItems[0]!;
};
