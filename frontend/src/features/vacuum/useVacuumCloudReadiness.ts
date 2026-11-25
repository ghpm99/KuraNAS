import { useVacuumCloudSettingsQuery } from './useVacuumCloudSettingsQuery';

export const useVacuumCloudReadiness = () => {
    const cloudSettingsQuery = useVacuumCloudSettingsQuery();
    const cloudSettings = cloudSettingsQuery.data;
    const isConfigured =
        Boolean(cloudSettings?.access_id) &&
        Boolean(cloudSettings?.region) &&
        cloudSettings?.has_access_secret === true;

    return {
        isChecking: cloudSettingsQuery.isPending,
        isConfigured,
        loadError: cloudSettingsQuery.isError ? cloudSettingsQuery.error : undefined,
        hasLoadError: cloudSettingsQuery.isError,
        refetch: () => void cloudSettingsQuery.refetch(),
    };
};
