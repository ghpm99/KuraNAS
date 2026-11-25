import { Alert, Button, CircularProgress } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useVacuumCloudConnectionTest } from '@/features/vacuum/useVacuumCloudConnectionTest';
import styles from './VacuumSettingsSection.module.css';

type VacuumCloudTestActionProps = {
    blockedMessageKey?: string;
};

const VacuumCloudTestAction = ({ blockedMessageKey }: VacuumCloudTestActionProps) => {
    const { t } = useI18n();
    const { outcome, isTesting, testCloud } = useVacuumCloudConnectionTest();
    const isBlocked = blockedMessageKey !== undefined;

    return (
        <div className={styles.fields}>
            <div className={styles.connectionRow}>
                <Button
                    variant="outlined"
                    className={styles.touchTarget}
                    onClick={testCloud}
                    disabled={isBlocked || isTesting}
                    startIcon={isTesting ? <CircularProgress size={16} role="progressbar" /> : null}
                >
                    {t('VACUUM_CLOUD_SETTINGS_TEST')}
                </Button>
            </div>
            {isBlocked ? <p className={styles.connectionHint}>{t(blockedMessageKey)}</p> : null}
            {outcome ? (
                <Alert severity={outcome.isSuccess ? 'success' : 'error'}>{outcome.message}</Alert>
            ) : null}
        </div>
    );
};

export default VacuumCloudTestAction;
