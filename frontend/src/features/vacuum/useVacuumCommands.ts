import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useRef } from 'react';
import { sendVacuumCommand } from '@/service/vacuum';
import type { VacuumCommand } from '@/types/vacuum';
import { toVacuumState } from './vacuumStateCatalog';
import { useVacuumFeedback } from './useVacuumFeedback';
import { useVacuumStatus } from './useVacuumStatus';
import { vacuumQueryKeys } from './vacuumQueryKeys';

export const useVacuumCommands = () => {
    const queryClient = useQueryClient();
    const { notifySuccess, notifyFailure } = useVacuumFeedback();
    const statusQuery = useVacuumStatus();
    const inFlightCommandRef = useRef<VacuumCommand | null>(null);

    const commandMutation = useMutation({
        mutationFn: (command: VacuumCommand) => sendVacuumCommand(command),
        onSuccess: () => notifySuccess('VACUUM_COMMAND_SENT'),
        onError: notifyFailure,
        onSettled: () => {
            inFlightCommandRef.current = null;
            return queryClient.invalidateQueries({ queryKey: vacuumQueryKeys.status });
        },
    });

    const sendCommand = (command: VacuumCommand) => {
        if (inFlightCommandRef.current === command) return;
        inFlightCommandRef.current = command;
        commandMutation.mutate(command);
    };

    return {
        state: toVacuumState(statusQuery.data?.state),
        pendingCommand: commandMutation.isPending ? commandMutation.variables : undefined,
        sendCommand,
    };
};
