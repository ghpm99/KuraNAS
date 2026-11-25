import type { VacuumCloudRegion } from '@/types/vacuum';

export const vacuumCloudRegionOptions: ReadonlyArray<{
    region: VacuumCloudRegion;
    labelKey: string;
}> = [
    { region: 'us', labelKey: 'VACUUM_CLOUD_SETTINGS_REGION_US' },
    { region: 'eu', labelKey: 'VACUUM_CLOUD_SETTINGS_REGION_EU' },
    { region: 'cn', labelKey: 'VACUUM_CLOUD_SETTINGS_REGION_CN' },
    { region: 'in', labelKey: 'VACUUM_CLOUD_SETTINGS_REGION_IN' },
];
