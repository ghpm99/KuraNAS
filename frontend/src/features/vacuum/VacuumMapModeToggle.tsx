import { ToggleButton, ToggleButtonGroup } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from './VacuumMapModeToggle.module.css';
import type { VacuumMapTouchMode } from './useVacuumMapTouchMode';

type VacuumMapModeToggleProps = {
    touchMode: VacuumMapTouchMode;
    onChange: (touchMode: VacuumMapTouchMode) => void;
};

const VacuumMapModeToggle = ({ touchMode, onChange }: VacuumMapModeToggleProps) => {
    const { t } = useI18n();

    return (
        <ToggleButtonGroup
            exclusive
            fullWidth
            size="small"
            value={touchMode}
            aria-label={t('VACUUM_MAP_MODE_LABEL')}
            onChange={(_event, nextMode: VacuumMapTouchMode | null) => {
                if (nextMode !== null) onChange(nextMode);
            }}
        >
            <ToggleButton className={styles.mode} value="select-rooms">
                {t('VACUUM_MAP_MODE_SELECT_ROOMS')}
            </ToggleButton>
            <ToggleButton className={styles.mode} value="clean-spot">
                {t('VACUUM_MAP_MODE_CLEAN_SPOT')}
            </ToggleButton>
        </ToggleButtonGroup>
    );
};

export default VacuumMapModeToggle;
