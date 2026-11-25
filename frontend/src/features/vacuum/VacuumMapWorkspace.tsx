import { useMemo } from 'react';
import type { VacuumRoom } from '@/types/vacuum';
import VacuumMapCanvas from './VacuumMapCanvas';
import VacuumMapCleaningActions from './VacuumMapCleaningActions';
import VacuumMapLegend from './VacuumMapLegend';
import VacuumMapModeToggle from './VacuumMapModeToggle';
import styles from './VacuumMapWorkspace.module.css';
import VacuumRoomList from './VacuumRoomList';
import VacuumRoomRenameDialog from './VacuumRoomRenameDialog';
import { useVacuumMapCleaning } from './useVacuumMapCleaning';
import { useVacuumMapTouchMode } from './useVacuumMapTouchMode';
import { useVacuumRoomRename } from './useVacuumRoomRename';
import { useVacuumRoomSelection } from './useVacuumRoomSelection';
import { useVacuumSpotPoint } from './useVacuumSpotPoint';
import type { MillimeterPoint } from './vacuumMapGeometry';
import type { VacuumMapModel } from './vacuumMapModel';

type VacuumMapWorkspaceProps = {
    map: VacuumMapModel;
    rooms: readonly VacuumRoom[];
    pathPoints: readonly (readonly [number, number])[];
    robotPosition: MillimeterPoint | null;
    canStartCleaning: boolean;
};

const VacuumMapWorkspace = ({
    map,
    rooms,
    pathPoints,
    robotPosition,
    canStartCleaning,
}: VacuumMapWorkspaceProps) => {
    const availableRoomIds = useMemo(() => rooms.map((room) => room.id), [rooms]);
    const { selectedRoomIds, toggleRoom, clearSelection } = useVacuumRoomSelection(
        map.mapId,
        availableRoomIds
    );
    const selectedRoomIdSet = useMemo(() => new Set(selectedRoomIds), [selectedRoomIds]);
    const { spotPoint, isSpotRejected, chooseSpot, clearSpot } = useVacuumSpotPoint(map);
    const { touchMode, changeTouchMode, tapMap } = useVacuumMapTouchMode(map, {
        onToggleRoom: toggleRoom,
        onChooseSpot: chooseSpot,
        onLeaveSpotMode: clearSpot,
    });
    const { roomBeingRenamed, isRenaming, askToRename, cancelRename, confirmRename } =
        useVacuumRoomRename(map.mapId);
    const { isSending, cleanRooms, cleanSpot } = useVacuumMapCleaning(map.mapId, {
        onCleaningStarted: () => {
            clearSelection();
            clearSpot();
        },
        onMapChanged: () => {
            clearSelection();
            clearSpot();
        },
    });

    return (
        <div className={styles.workspace}>
            <div className={styles.mapColumn}>
                <VacuumMapModeToggle touchMode={touchMode} onChange={changeTouchMode} />
                <VacuumMapCanvas
                    map={map}
                    selectedRoomIds={selectedRoomIdSet}
                    pathPoints={pathPoints}
                    robotPosition={robotPosition}
                    spotCenter={touchMode === 'clean-spot' ? spotPoint : null}
                    onTap={tapMap}
                />
                <VacuumMapCleaningActions
                    touchMode={touchMode}
                    selectedRoomCount={selectedRoomIds.length}
                    hasSpotPoint={spotPoint !== null}
                    isSpotRejected={isSpotRejected}
                    canStartCleaning={canStartCleaning}
                    isSending={isSending}
                    onCleanRooms={() => cleanRooms([...selectedRoomIds])}
                    onCleanSpot={() => {
                        if (spotPoint !== null) cleanSpot(spotPoint);
                    }}
                />
                <VacuumMapLegend />
            </div>
            <div className={styles.roomsColumn}>
                <VacuumRoomList
                    rooms={rooms}
                    selectedRoomIds={selectedRoomIds}
                    onToggleRoom={toggleRoom}
                    onRenameRoom={askToRename}
                />
            </div>
            <VacuumRoomRenameDialog
                room={roomBeingRenamed}
                isSaving={isRenaming}
                onCancel={cancelRename}
                onConfirm={confirmRename}
            />
        </div>
    );
};

export default VacuumMapWorkspace;
