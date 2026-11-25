import { useQuery } from '@tanstack/react-query';
import { getVacuumMap } from '@/service/vacuum';
import { useVacuumState } from './useVacuumState';
import { toVacuumMapModel } from './vacuumMapModel';
import { vacuumQueryKeys } from './vacuumQueryKeys';

export const vacuumMapPollingIntervalMs = 30000;

export const useVacuumMap = (isEnabled: boolean) => {
    const isCleaning = useVacuumState() === 'cleaning';
    const mapQuery = useQuery({
        queryKey: vacuumQueryKeys.map,
        queryFn: getVacuumMap,
        enabled: isEnabled,
        retry: false,
        refetchInterval: isCleaning ? vacuumMapPollingIntervalMs : false,
        refetchIntervalInBackground: false,
    });

    return { mapQuery, map: toVacuumMapModel(mapQuery.data) };
};
