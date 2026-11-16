import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useRef, useState } from 'react';
import { getVacuumConsumables, resetVacuumConsumable } from '@/service/vacuum';
import type { VacuumConsumablePart } from '@/types/vacuum';
import { useVacuumFeedback } from './useVacuumFeedback';
import { vacuumQueryKeys } from './vacuumQueryKeys';

export const useVacuumConsumables = () => {
    const queryClient = useQueryClient();
    const { notifySuccess, notifyFailure } = useVacuumFeedback();
    const [partAwaitingConfirmation, setPartAwaitingConfirmation] =
        useState<VacuumConsumablePart | null>(null);

    const isResetInFlightRef = useRef(false);

    const consumablesQuery = useQuery({
        queryKey: vacuumQueryKeys.consumables,
        queryFn: getVacuumConsumables,
        retry: false,
    });

    const resetMutation = useMutation({
        mutationFn: (part: VacuumConsumablePart) => resetVacuumConsumable(part),
        onSuccess: () => notifySuccess('VACUUM_RESET_DONE'),
        onError: notifyFailure,
        onSettled: () => {
            isResetInFlightRef.current = false;
            setPartAwaitingConfirmation(null);
            return queryClient.invalidateQueries({ queryKey: vacuumQueryKeys.consumables });
        },
    });

    const confirmReset = () => {
        if (partAwaitingConfirmation === null || isResetInFlightRef.current) return;
        isResetInFlightRef.current = true;
        resetMutation.mutate(partAwaitingConfirmation);
    };

    return {
        consumablesQuery,
        partAwaitingConfirmation,
        isResetting: resetMutation.isPending,
        askToReset: setPartAwaitingConfirmation,
        cancelReset: () => setPartAwaitingConfirmation(null),
        confirmReset,
    };
};
