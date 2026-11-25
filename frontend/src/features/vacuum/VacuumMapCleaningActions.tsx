import { Alert, Button } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from './VacuumMapCleaningActions.module.css';
import type { VacuumMapTouchMode } from './useVacuumMapTouchMode';

type VacuumMapCleaningActionsProps = {
    touchMode: VacuumMapTouchMode;
    selectedRoomCount: number;
    hasSpotPoint: boolean;
    isSpotRejected: boolean;
    canStartCleaning: boolean;
    isSending: boolean;
    onCleanRooms: () => void;
    onCleanSpot: () => void;
};

const resolveHintKey = (
    touchMode: VacuumMapTouchMode,
    hasSpotPoint: boolean,
    selectedRoomCount: number
): string | null => {
    if (touchMode === 'select-rooms') {
        return selectedRoomCount === 0 ? 'VACUUM_MAP_ROOMS_PROMPT' : null;
    }
    return hasSpotPoint ? 'VACUUM_MAP_SPOT_CONFIRM' : 'VACUUM_MAP_SPOT_PROMPT';
};

const VacuumMapCleaningActions = ({
    touchMode,
    selectedRoomCount,
    hasSpotPoint,
    isSpotRejected,
    canStartCleaning,
    isSending,
    onCleanRooms,
    onCleanSpot,
}: VacuumMapCleaningActionsProps) => {
    const { t } = useI18n();
    const hintKey = resolveHintKey(touchMode, hasSpotPoint, selectedRoomCount);
    const isSpotMode = touchMode === 'clean-spot';

    return (
        <div className={styles.actions}>
            {hintKey && <p className={styles.hint}>{t(hintKey)}</p>}
            {isSpotMode && isSpotRejected && (
                <Alert severity="warning" role="status">
                    {t('VACUUM_MAP_SPOT_REJECTED')}
                </Alert>
            )}
            {!canStartCleaning && (
                <p className={styles.hint}>{t('VACUUM_MAP_ACTIONS_UNAVAILABLE')}</p>
            )}
            <div className={styles.buttons}>
                {selectedRoomCount > 0 && (
                    <Button
                        className={styles.button}
                        variant="contained"
                        disabled={!canStartCleaning || isSending}
                        onClick={onCleanRooms}
                    >
                        {t('VACUUM_MAP_CLEAN_SELECTED', { count: String(selectedRoomCount) })}
                    </Button>
                )}
                {isSpotMode && hasSpotPoint && (
                    <Button
                        className={styles.button}
                        variant="contained"
                        disabled={!canStartCleaning || isSending}
                        onClick={onCleanSpot}
                    >
                        {t('VACUUM_MAP_CLEAN_SPOT')}
                    </Button>
                )}
            </div>
        </div>
    );
};

export default VacuumMapCleaningActions;
