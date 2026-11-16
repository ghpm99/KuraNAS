import { useQuery } from '@tanstack/react-query';
import { getVacuumTotals } from '@/service/vacuum';
import { vacuumQueryKeys } from './vacuumQueryKeys';

export const useVacuumTotals = () =>
    useQuery({
        queryKey: vacuumQueryKeys.totals,
        queryFn: getVacuumTotals,
        retry: false,
    });
