import { Alert, Button, CircularProgress } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useVacuumConnectionActions } from '@/features/vacuum/useVacuumConnectionActions';
import styles from './VacuumSettingsSection.module.css';

type VacuumConnectionActionsProps = {
    blockedMessageKey?: string;
};

const VacuumConnectionActions = ({ blockedMessageKey }: VacuumConnectionActionsProps) => {
    const { t } = useI18n();
    const { outcome, isDiscovering, isTesting, discover, testConnection } =
        useVacuumConnectionActions();
    const isBlocked = blockedMessageKey !== undefined;

    return (
        <div className={styles.fields}>
            <div className={styles.connectionRow}>
                <Button
                    variant="outlined"
                    onClick={discover}
                    disabled={isBlocked || isDiscovering}
                    startIcon={isDiscovering ? <CircularProgress size={16} role="progressbar" /> : null}
                >
                    {t('VACUUM_SETTINGS_DISCOVER')}
                </Button>
                <Button
                    variant="outlined"
                    onClick={testConnection}
                    disabled={isBlocked || isTesting}
                    startIcon={isTesting ? <CircularProgress size={16} role="progressbar" /> : null}
                >
                    {t('VACUUM_SETTINGS_TEST')}
                </Button>
            </div>
            {isBlocked ? <p className={styles.connectionHint}>{t(blockedMessageKey)}</p> : null}
            {outcome ? (
                <Alert severity={outcome.isSuccess ? 'success' : 'error'}>{outcome.message}</Alert>
            ) : null}
        </div>
    );
};

export default VacuumConnectionActions;
