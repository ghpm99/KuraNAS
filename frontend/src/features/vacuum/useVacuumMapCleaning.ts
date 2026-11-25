import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useRef } from 'react';
import { startVacuumRoomCleaning, startVacuumSpotCleaning } from '@/service/vacuum';
import type { MillimeterPoint } from './vacuumMapGeometry';
import { useVacuumFeedback } from './useVacuumFeedback';
import { useVacuumMapFailureHandler } from './useVacuumMapFailureHandler';
import { vacuumQueryKeys } from './vacuumQueryKeys';

type MapCleaningCallbacks = {
    onCleaningStarted: () => void;
    onMapChanged: () => void;
};

export const useVacuumMapCleaning = (
    mapId: number | undefined,
    { onCleaningStarted, onMapChanged }: MapCleaningCallbacks
) => {
    const queryClient = useQueryClient();
    const { notifySuccess } = useVacuumFeedback();
    const handleFailure = useVacuumMapFailureHandler(onMapChanged);
    const isSendingRef = useRef(false);

    const handleSettled = () => {
        isSendingRef.current = false;
        void queryClient.invalidateQueries({ queryKey: vacuumQueryKeys.status });
        void queryClient.invalidateQueries({ queryKey: vacuumQueryKeys.path });
    };

    const handleSuccess = () => {
        notifySuccess('VACUUM_MAP_CLEANING_SENT');
        onCleaningStarted();
    };

    const roomCleaningMutation = useMutation({
        mutationFn: (roomIds: number[]) =>
            startVacuumRoomCleaning({ map_id: mapId ?? 0, room_ids: roomIds }),
        onSuccess: handleSuccess,
        onError: handleFailure,
        onSettled: handleSettled,
    });

    const spotCleaningMutation = useMutation({
        mutationFn: (point: MillimeterPoint) =>
            startVacuumSpotCleaning({ map_id: mapId ?? 0, x: point.x, y: point.y }),
        onSuccess: handleSuccess,
        onError: handleFailure,
        onSettled: handleSettled,
    });

    const cleanRooms = (roomIds: number[]) => {
        if (isSendingRef.current || mapId === undefined) return;
        isSendingRef.current = true;
        roomCleaningMutation.mutate(roomIds);
    };

    const cleanSpot = (point: MillimeterPoint) => {
        if (isSendingRef.current || mapId === undefined) return;
        isSendingRef.current = true;
        spotCleaningMutation.mutate(point);
    };

    return {
        isSending: roomCleaningMutation.isPending || spotCleaningMutation.isPending,
        cleanRooms,
        cleanSpot,
    };
};
