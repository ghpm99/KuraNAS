import type { VacuumSettings } from '@/types/vacuum';

export type SavedVacuumSettings = {
    address: string;
    deviceId: string;
    hasLocalKey: boolean;
};

export const toSavedVacuumSettings = (
    settings: Partial<VacuumSettings> | null | undefined
): SavedVacuumSettings => ({
    address: settings?.address ?? '',
    deviceId: settings?.device_id ?? '',
    hasLocalKey: settings?.has_local_key === true,
});
