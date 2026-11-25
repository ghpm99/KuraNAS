import { Checkbox, FormControlLabel, IconButton } from '@mui/material';
import { Pencil } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { VacuumRoom } from '@/types/vacuum';
import { formatVacuumRoomArea } from './formatVacuumRoomArea';
import { resolveVacuumRoomName } from './resolveVacuumRoomName';
import styles from './VacuumRoomList.module.css';

type VacuumRoomListProps = {
    rooms: readonly VacuumRoom[];
    selectedRoomIds: readonly number[];
    onToggleRoom: (roomId: number) => void;
    onRenameRoom: (room: VacuumRoom) => void;
};

const VacuumRoomList = ({
    rooms,
    selectedRoomIds,
    onToggleRoom,
    onRenameRoom,
}: VacuumRoomListProps) => {
    const { t } = useI18n();
    const decimalSeparator = t('VACUUM_ROOM_AREA_DECIMAL_SEPARATOR');

    return (
        <div className={styles.rooms}>
            <h3 className={styles.title}>{t('VACUUM_ROOMS_TITLE')}</h3>
            {rooms.length === 0 && <p className={styles.empty}>{t('VACUUM_ROOMS_EMPTY')}</p>}
            <ul className={styles.list}>
                {rooms.map((room) => {
                    const roomName = resolveVacuumRoomName(room, t);
                    return (
                        <li key={room.id} className={styles.room}>
                            <FormControlLabel
                                className={styles.selection}
                                control={
                                    <Checkbox
                                        className={styles.checkbox}
                                        checked={selectedRoomIds.includes(room.id)}
                                        onChange={() => onToggleRoom(room.id)}
                                    />
                                }
                                label={
                                    <span className={styles.description}>
                                        <span className={styles.name}>{roomName}</span>
                                        <span className={styles.area}>
                                            {t('VACUUM_ROOM_AREA', {
                                                area: formatVacuumRoomArea(room.area, decimalSeparator),
                                            })}
                                        </span>
                                    </span>
                                }
                            />
                            <IconButton
                                className={styles.rename}
                                aria-label={t('VACUUM_ROOM_RENAME_LABEL', { room: roomName })}
                                onClick={() => onRenameRoom(room)}
                            >
                                <Pencil size={18} />
                            </IconButton>
                        </li>
                    );
                })}
            </ul>
        </div>
    );
};

export default VacuumRoomList;
