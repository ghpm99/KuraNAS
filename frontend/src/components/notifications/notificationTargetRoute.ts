import { appRoutes } from '@/app/routes';
import type { Notification } from '@/types/notification';

const eventPrefixRoutes: ReadonlyArray<readonly [string, string]> = [
    ['capture_', appRoutes.captures],
    ['takeout_', appRoutes.takeout],
    ['watch_folder_', appRoutes.files],
];

const groupKeyPrefixRoutes: ReadonlyArray<readonly [string, string]> = [
    ['capture_', appRoutes.captures],
    ['takeout_', appRoutes.takeout],
    ['watch_import_', appRoutes.files],
    ['ytdlp-', appRoutes.settings],
];

const findRouteByPrefix = (
    value: string,
    prefixRoutes: ReadonlyArray<readonly [string, string]>
): string | undefined => prefixRoutes.find(([prefix]) => value.startsWith(prefix))?.[1];

export const resolveNotificationTargetRoute = (notification: Notification): string | undefined => {
    const event = notification.metadata?.event;
    if (typeof event === 'string') {
        const routeFromEvent = findRouteByPrefix(event, eventPrefixRoutes);
        if (routeFromEvent) return routeFromEvent;
    }
    if (notification.group_key) {
        return findRouteByPrefix(notification.group_key, groupKeyPrefixRoutes);
    }
    return undefined;
};
