import type { VacuumCloudSettings } from '@/types/vacuum';

export type SavedVacuumCloudSettings = {
    accessId: string;
    region: string;
    hasAccessSecret: boolean;
};

export const toSavedVacuumCloudSettings = (
    settings: Partial<VacuumCloudSettings> | null | undefined
): SavedVacuumCloudSettings => ({
    accessId: settings?.access_id ?? '',
    region: settings?.region ?? '',
    hasAccessSecret: settings?.has_access_secret === true,
});
