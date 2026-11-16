import { useQuery } from '@tanstack/react-query';
import { getVacuumSettings } from '@/service/vacuum';
import { vacuumQueryKeys } from './vacuumQueryKeys';

export const useVacuumSettingsQuery = () =>
    useQuery({
        queryKey: vacuumQueryKeys.settings,
        queryFn: getVacuumSettings,
        retry: false,
    });
