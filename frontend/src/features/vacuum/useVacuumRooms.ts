import { useQuery } from '@tanstack/react-query';
import { getVacuumRooms } from '@/service/vacuum';
import type { VacuumRoom } from '@/types/vacuum';
import { useVacuumState } from './useVacuumState';
import { vacuumMapPollingIntervalMs } from './useVacuumMap';
import { vacuumQueryKeys } from './vacuumQueryKeys';

const noRooms: VacuumRoom[] = [];

export const useVacuumRooms = (isEnabled: boolean) => {
    const isCleaning = useVacuumState() === 'cleaning';
    const roomsQuery = useQuery({
        queryKey: vacuumQueryKeys.rooms,
        queryFn: getVacuumRooms,
        enabled: isEnabled,
        retry: false,
        refetchInterval: isCleaning ? vacuumMapPollingIntervalMs : false,
        refetchIntervalInBackground: false,
    });
    const reportedRooms = roomsQuery.data?.items;

    return {
        roomsQuery,
        rooms: Array.isArray(reportedRooms) ? reportedRooms : noRooms,
    };
};
