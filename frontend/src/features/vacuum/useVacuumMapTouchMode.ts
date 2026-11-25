import { useCallback, useState } from 'react';
import type { MillimeterPoint } from './vacuumMapGeometry';
import { findRoomIdAt } from './vacuumMapGeometry';
import type { VacuumMapModel } from './vacuumMapModel';

export type VacuumMapTouchMode = 'select-rooms' | 'clean-spot';

type TouchModeCallbacks = {
    onToggleRoom: (roomId: number) => void;
    onChooseSpot: (point: MillimeterPoint) => void;
    onLeaveSpotMode: () => void;
};

export const useVacuumMapTouchMode = (
    map: VacuumMapModel | null,
    { onToggleRoom, onChooseSpot, onLeaveSpotMode }: TouchModeCallbacks
) => {
    const [touchMode, setTouchMode] = useState<VacuumMapTouchMode>('select-rooms');

    const changeTouchMode = useCallback(
        (nextMode: VacuumMapTouchMode) => {
            setTouchMode(nextMode);
            if (nextMode === 'select-rooms') onLeaveSpotMode();
        },
        [onLeaveSpotMode]
    );

    const tapMap = useCallback(
        (point: MillimeterPoint) => {
            if (map === null) return;
            if (touchMode === 'clean-spot') {
                onChooseSpot(point);
                return;
            }
            const roomId = findRoomIdAt(map, point);
            if (roomId !== null) onToggleRoom(roomId);
        },
        [map, touchMode, onChooseSpot, onToggleRoom]
    );

    return { touchMode, changeTouchMode, tapMap };
};
