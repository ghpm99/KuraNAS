import { Alert, Button, CircularProgress } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import { toSavedVacuumSettings } from '@/features/vacuum/toSavedVacuumSettings';
import { useVacuumSettingsQuery } from '@/features/vacuum/useVacuumSettingsQuery';
import { resolveVacuumFailureMessage } from '@/features/vacuum/vacuumFailures';
import VacuumSettingsForm from './VacuumSettingsForm';
import styles from './VacuumSettingsSection.module.css';

type VacuumSettingsSectionProps = {
    className?: string;
};

const VacuumSettingsSection = ({ className = '' }: VacuumSettingsSectionProps) => {
    const { t } = useI18n();
    const settingsQuery = useVacuumSettingsQuery();
    const isLoadFailed = settingsQuery.isError && settingsQuery.data === undefined;
    const isFormVisible = !settingsQuery.isPending && !isLoadFailed;

    return (
        <section className={`${className} ${styles.section}`.trim()}>
            <div className={styles.header}>
                <h2 className={styles.title}>{t('VACUUM_SETTINGS_TITLE')}</h2>
                <p className={styles.description}>{t('VACUUM_SETTINGS_DESCRIPTION')}</p>
            </div>
            {settingsQuery.isPending ? <CircularProgress size={20} /> : null}
            {isLoadFailed ? (
                <Alert
                    severity="error"
                    action={
                        <Button color="inherit" size="small" onClick={() => void settingsQuery.refetch()}>
                            {t('TRY_AGAIN')}
                        </Button>
                    }
                >
                    {resolveVacuumFailureMessage(settingsQuery.error, t('VACUUM_SETTINGS_LOAD_ERROR'))}
                </Alert>
            ) : null}
            {isFormVisible ? (
                <VacuumSettingsForm savedSettings={toSavedVacuumSettings(settingsQuery.data)} />
            ) : null}
        </section>
    );
};

export default VacuumSettingsSection;
