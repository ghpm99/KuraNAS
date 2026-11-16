import { useQuery } from '@tanstack/react-query';
import { getVacuumSettings } from '@/service/vacuum';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { isVacuumNotConfiguredError } from './vacuumFailures';
import { vacuumQueryKeys } from './vacuumQueryKeys';
import { useVacuumStatus } from './useVacuumStatus';

export const useVacuumAvailability = () => {
    const settingsQuery = useQuery({
        queryKey: vacuumQueryKeys.settings,
        queryFn: getVacuumSettings,
        retry: false,
    });
    const settings = settingsQuery.data;
    const isDeviceIdMissing = settings !== undefined && !settings.device_id;
    const isLocalKeyMissing = settings !== undefined && !settings.has_local_key;
    const isSettingsIncomplete = isDeviceIdMissing || isLocalKeyMissing;

    const statusQuery = useVacuumStatus(!settingsQuery.isPending && !isSettingsIncomplete);
    const isRejectedAsNotConfigured = isVacuumNotConfiguredError(statusQuery.error);

    return {
        isChecking: settingsQuery.isPending,
        isNotConfigured: isSettingsIncomplete || isRejectedAsNotConfigured,
        isDeviceIdMissing,
        isLocalKeyMissing,
        serverMessage: isRejectedAsNotConfigured
            ? extractBackendErrorMessage(statusQuery.error)
            : undefined,
    };
};
