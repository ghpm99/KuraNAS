import { useQuery } from '@tanstack/react-query';
import { getVacuumCloudSettings } from '@/service/vacuum';
import { vacuumQueryKeys } from './vacuumQueryKeys';

export const useVacuumCloudSettingsQuery = () =>
    useQuery({
        queryKey: vacuumQueryKeys.cloudSettings,
        queryFn: getVacuumCloudSettings,
        retry: false,
    });
