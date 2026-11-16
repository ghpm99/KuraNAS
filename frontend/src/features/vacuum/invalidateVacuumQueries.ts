import type { QueryClient } from '@tanstack/react-query';
import { vacuumQueryKeys } from './vacuumQueryKeys';

export const invalidateVacuumQueries = (queryClient: QueryClient): Promise<unknown> =>
    Promise.all(
        Object.values(vacuumQueryKeys).map((queryKey) => queryClient.invalidateQueries({ queryKey }))
    );
