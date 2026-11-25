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

    it('lists documents right after videos with its own translation keys', () => {
        const itemRoutes = navigationItems.map((item) => item.href);
        const documentsPosition = itemRoutes.indexOf(appRoutes.documents);

        expect(itemRoutes[documentsPosition - 1]).toBe(appRoutes.videos);
        expect(navigationItems[documentsPosition]?.labelKey).toBe('NAV_DOCUMENTS');
        expect(navigationItems[documentsPosition]?.descriptionKey).toBe('DOCUMENTS_PAGE_DESCRIPTION');
    });

    it('lists the download from URL tool right before the captures with its own translation keys', () => {
        const itemRoutes = navigationItems.map((item) => item.href);
        const ingestPosition = itemRoutes.indexOf(appRoutes.ingest);

        expect(itemRoutes[ingestPosition + 1]).toBe(appRoutes.captures);
        expect(navigationItems[ingestPosition]?.labelKey).toBe('NAV_INGEST');
        expect(navigationItems[ingestPosition]?.descriptionKey).toBe('INGEST_PAGE_DESCRIPTION');
    });

    it('has no repeated destination', () => {
        const itemRoutes = navigationItems.map((item) => item.href);

        expect(new Set(itemRoutes).size).toBe(itemRoutes.length);
    });
});
