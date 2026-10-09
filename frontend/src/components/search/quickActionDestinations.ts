import {
    appRoutes,
    getAnalyticsRoute,
    getImageRoute,
    getMusicRoute,
    getVideoRoute,
} from '@/app/routes';
import { analyticsNavigationItems } from '@/components/analytics/navigation';
import { imageNavigationItems } from '@/components/images/navigation';
import { navigationItems } from '@/components/layout/navigationItems';
import { getSettingsSectionRoute, settingsSections } from '@/components/settings/settingsSections';
import { musicNavigationItems } from '@/features/music/components/navigation';
import { videoNavigationItems } from '@/features/videos/components/navigation';

export type QuickActionDestination = {
    route: string;
    labelKey: string;
    descriptionKey: string;
    parentLabelKey?: string;
};

type DomainSection<SectionKey> = {
    key: SectionKey;
    labelKey: string;
    descriptionKey: string;
};

const toSectionDestinations = <SectionKey>(
    parentLabelKey: string,
    sections: DomainSection<SectionKey>[],
    getRoute: (sectionKey: SectionKey) => string
): QuickActionDestination[] =>
    sections.map((section) => ({
        route: getRoute(section.key),
        labelKey: section.labelKey,
        descriptionKey: section.descriptionKey,
        parentLabelKey,
    }));

const sidebarDestinations: QuickActionDestination[] = navigationItems.map((item) => ({
    route: item.href,
    labelKey: item.labelKey,
    descriptionKey: item.descriptionKey,
}));

const routedPagesOutsideSidebar: QuickActionDestination[] = [
    {
        route: appRoutes.notifications,
        labelKey: 'NOTIFICATIONS',
        descriptionKey: 'GLOBAL_SEARCH_ACTION_NOTIFICATIONS_DESCRIPTION',
    },
    {
        route: appRoutes.activityDiary,
        labelKey: 'ACTIVITY_DIARY',
        descriptionKey: 'ACTIVITY_DIARY_SUBTITLE',
    },
];

const musicSearchDestination: QuickActionDestination = {
    route: getMusicRoute('search'),
    labelKey: 'GLOBAL_SEARCH_ACTION_MUSIC_SEARCH',
    descriptionKey: 'MUSIC_SEARCH_PROMPT',
    parentLabelKey: 'NAV_MUSIC',
};

const domainSectionDestinations: QuickActionDestination[] = [
    ...toSectionDestinations('NAV_IMAGES', imageNavigationItems, getImageRoute),
    ...toSectionDestinations('NAV_MUSIC', musicNavigationItems, getMusicRoute),
    musicSearchDestination,
    ...toSectionDestinations('NAV_VIDEOS', videoNavigationItems, getVideoRoute),
    ...toSectionDestinations('ANALYTICS', analyticsNavigationItems, getAnalyticsRoute),
];

const settingsSectionDestinations: QuickActionDestination[] = settingsSections.map((section) => ({
    route: getSettingsSectionRoute(section.id),
    labelKey: section.labelKey,
    descriptionKey: section.descriptionKey,
    parentLabelKey: 'SETTINGS',
}));

const keepFirstDestinationPerRoute = (destinations: QuickActionDestination[]) => {
    const seenRoutes = new Set<string>();
    return destinations.filter((destination) => {
        if (seenRoutes.has(destination.route)) {
            return false;
        }
        seenRoutes.add(destination.route);
        return true;
    });
};

export const quickActionDestinations: QuickActionDestination[] = keepFirstDestinationPerRoute([
    ...sidebarDestinations,
    ...routedPagesOutsideSidebar,
    ...domainSectionDestinations,
    ...settingsSectionDestinations,
]);
