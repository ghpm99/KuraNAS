import { useQuery } from '@tanstack/react-query';
import { getVacuumStatus } from '@/service/vacuum';
import { isVacuumNotConfiguredError } from './vacuumFailures';
import { vacuumQueryKeys } from './vacuumQueryKeys';

export const vacuumStatusPollingIntervalMs = 5000;

export const useVacuumStatus = (isEnabled = true) =>
    useQuery({
        queryKey: vacuumQueryKeys.status,
        queryFn: getVacuumStatus,
        enabled: isEnabled,
        retry: false,
        refetchInterval: (query) =>
            isVacuumNotConfiguredError(query.state.error) ? false : vacuumStatusPollingIntervalMs,
        refetchIntervalInBackground: false,
    });
