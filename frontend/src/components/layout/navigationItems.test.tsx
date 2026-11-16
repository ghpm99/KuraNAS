import { appRoutes } from '@/app/routes';
import { navigationItems } from './navigationItems';

describe('navigationItems', () => {
    it('lists the robot vacuum right before the settings', () => {
        const itemRoutes = navigationItems.map((item) => item.href);
        const vacuumPosition = itemRoutes.indexOf(appRoutes.vacuum);

        expect(vacuumPosition).toBeGreaterThan(-1);
        expect(itemRoutes[vacuumPosition + 1]).toBe(appRoutes.settings);
    });

    it('labels the robot vacuum item with its own translation keys', () => {
        const vacuumItem = navigationItems.find((item) => item.href === appRoutes.vacuum);

        expect(vacuumItem?.labelKey).toBe('NAV_VACUUM');
        expect(vacuumItem?.descriptionKey).toBe('VACUUM_PAGE_DESCRIPTION');
    });

    it('has no repeated destination', () => {
        const itemRoutes = navigationItems.map((item) => item.href);

        expect(new Set(itemRoutes).size).toBe(itemRoutes.length);
    });
});
