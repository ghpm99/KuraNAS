import { useCallback, useState } from 'react';
import type { VacuumMapModel } from './vacuumMapModel';
import { findRoomIdAt, type MillimeterPoint } from './vacuumMapGeometry';

type SpotOfMap = { mapId: number; point: MillimeterPoint | null; isRejected: boolean };

export const useVacuumSpotPoint = (map: VacuumMapModel | null) => {
    const [spot, setSpot] = useState<SpotOfMap | null>(null);
    const currentSpot = spot !== null && spot.mapId === map?.mapId ? spot : null;

    const chooseSpot = useCallback(
        (point: MillimeterPoint) => {
            if (map === null) return;
            const isOnFloor = findRoomIdAt(map, point) !== null;
            setSpot((current) => {
                const previousPoint = current?.mapId === map.mapId ? current.point : null;
                return {
                    mapId: map.mapId,
                    point: isOnFloor ? point : previousPoint,
                    isRejected: !isOnFloor,
                };
            });
        },
        [map]
    );

    const clearSpot = useCallback(() => setSpot(null), []);

    return {
        spotPoint: currentSpot?.point ?? null,
        isSpotRejected: currentSpot?.isRejected ?? false,
        chooseSpot,
        clearSpot,
    };
};
