import { useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';
import { useVacuumFeedback } from './useVacuumFeedback';
import { isVacuumNotConfiguredError } from './vacuumFailures';
import { vacuumQueryKeys } from './vacuumQueryKeys';

export const useVacuumMapFailureHandler = (onMapChanged?: () => void) => {
    const queryClient = useQueryClient();
    const { notifyFailure } = useVacuumFeedback();

    return useCallback(
        (error: unknown) => {
            notifyFailure(error);
            if (!isVacuumNotConfiguredError(error)) return;
            onMapChanged?.();
            void queryClient.invalidateQueries({ queryKey: vacuumQueryKeys.map });
            void queryClient.invalidateQueries({ queryKey: vacuumQueryKeys.rooms });
        },
        [notifyFailure, onMapChanged, queryClient]
    );
};
