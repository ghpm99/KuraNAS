import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useRef, useState } from 'react';
import { renameVacuumRoom } from '@/service/vacuum';
import type { VacuumRoom } from '@/types/vacuum';
import { useVacuumFeedback } from './useVacuumFeedback';
import { useVacuumMapFailureHandler } from './useVacuumMapFailureHandler';
import { vacuumQueryKeys } from './vacuumQueryKeys';

export const useVacuumRoomRename = (mapId: number | undefined) => {
    const queryClient = useQueryClient();
    const { notifySuccess } = useVacuumFeedback();
    const handleFailure = useVacuumMapFailureHandler();
    const [roomBeingRenamed, setRoomBeingRenamed] = useState<VacuumRoom | null>(null);
    const isRenamingRef = useRef(false);

    const renameMutation = useMutation({
        mutationFn: (rename: { roomId: number; name: string }) =>
            renameVacuumRoom(rename.roomId, { map_id: mapId ?? 0, name: rename.name }),
        onSuccess: () => {
            notifySuccess('VACUUM_ROOM_RENAMED');
            setRoomBeingRenamed(null);
        },
        onError: handleFailure,
        onSettled: () => {
            isRenamingRef.current = false;
            return queryClient.invalidateQueries({ queryKey: vacuumQueryKeys.rooms });
        },
    });

    const confirmRename = (name: string) => {
        if (roomBeingRenamed === null || mapId === undefined || isRenamingRef.current) return;
        isRenamingRef.current = true;
        renameMutation.mutate({ roomId: roomBeingRenamed.id, name });
    };

    return {
        roomBeingRenamed,
        isRenaming: renameMutation.isPending,
        askToRename: setRoomBeingRenamed,
        cancelRename: () => setRoomBeingRenamed(null),
        confirmRename,
    };
};
