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
import { quickActionDestinations } from './quickActionDestinations';

const destinationRoutes = quickActionDestinations.map((destination) => destination.route);

describe('search/quickActionDestinations', () => {
    it('contains every sidebar item', () => {
        navigationItems.forEach((item) => {
            expect(destinationRoutes).toContain(item.href);
        });
    });

    it('contains the sidebar destinations named by the product review', () => {
        [
            appRoutes.trash,
            appRoutes.downloads,
            appRoutes.captures,
            appRoutes.assistant,
            appRoutes.takeout,
            appRoutes.settings,
            appRoutes.about,
        ].forEach((route) => expect(destinationRoutes).toContain(route));
    });

    it('contains routed pages that are not in the sidebar', () => {
        expect(destinationRoutes).toContain(appRoutes.notifications);
        expect(destinationRoutes).toContain(appRoutes.activityDiary);
    });

    it('contains every images, music, videos and analytics section', () => {
        imageNavigationItems.forEach((item) =>
            expect(destinationRoutes).toContain(getImageRoute(item.key))
        );
        musicNavigationItems.forEach((item) =>
            expect(destinationRoutes).toContain(getMusicRoute(item.key))
        );
        videoNavigationItems.forEach((item) =>
            expect(destinationRoutes).toContain(getVideoRoute(item.key))
        );
        analyticsNavigationItems.forEach((item) =>
            expect(destinationRoutes).toContain(getAnalyticsRoute(item.key))
        );
        expect(destinationRoutes).toContain(getMusicRoute('search'));
    });

    it('deep links every settings section by its anchor id', () => {
        settingsSections.forEach((section) =>
            expect(destinationRoutes).toContain(getSettingsSectionRoute(section.id))
        );
    });

    it('never repeats a route', () => {
        expect(new Set(destinationRoutes).size).toBe(destinationRoutes.length);
    });

    it('gives every destination a label key and a description key', () => {
        quickActionDestinations.forEach((destination) => {
            expect(destination.labelKey).not.toBe('');
            expect(destination.descriptionKey).not.toBe('');
        });
    });
});
