import { useCallback, useMemo, useState } from 'react';

type SelectionOfMap = { mapId: number | undefined; roomIds: readonly number[] };

export const useVacuumRoomSelection = (mapId: number | undefined, availableRoomIds: readonly number[]) => {
    const [selection, setSelection] = useState<SelectionOfMap>({ mapId: undefined, roomIds: [] });

    const selectedRoomIds = useMemo(
        () =>
            selection.mapId === mapId
                ? selection.roomIds.filter((roomId) => availableRoomIds.includes(roomId))
                : [],
        [selection, mapId, availableRoomIds]
    );

    const toggleRoom = useCallback(
        (roomId: number) => {
            setSelection((current) => {
                const currentRoomIds = current.mapId === mapId ? current.roomIds : [];
                const nextRoomIds = currentRoomIds.includes(roomId)
                    ? currentRoomIds.filter((selectedRoomId) => selectedRoomId !== roomId)
                    : [...currentRoomIds, roomId];
                return { mapId, roomIds: nextRoomIds };
            });
        },
        [mapId]
    );

    const clearSelection = useCallback(() => setSelection({ mapId, roomIds: [] }), [mapId]);

    return { selectedRoomIds, toggleRoom, clearSelection };
};
