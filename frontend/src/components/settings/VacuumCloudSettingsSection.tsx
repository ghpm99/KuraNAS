import { Alert, Button, CircularProgress } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import { toSavedVacuumCloudSettings } from '@/features/vacuum/toSavedVacuumCloudSettings';
import { useVacuumCloudSettingsQuery } from '@/features/vacuum/useVacuumCloudSettingsQuery';
import { resolveVacuumFailureMessage } from '@/features/vacuum/vacuumFailures';
import VacuumCloudSettingsForm from './VacuumCloudSettingsForm';
import styles from './VacuumSettingsSection.module.css';

type VacuumCloudSettingsSectionProps = {
    className?: string;
};

const VacuumCloudSettingsSection = ({ className = '' }: VacuumCloudSettingsSectionProps) => {
    const { t } = useI18n();
    const settingsQuery = useVacuumCloudSettingsQuery();
    const isLoadFailed = settingsQuery.isError && settingsQuery.data === undefined;
    const isFormVisible = !settingsQuery.isPending && !isLoadFailed;

    return (
        <section className={`${className} ${styles.section}`.trim()}>
            <div className={styles.header}>
                <h2 className={styles.title}>{t('VACUUM_CLOUD_SETTINGS_TITLE')}</h2>
                <p className={styles.description}>{t('VACUUM_CLOUD_SETTINGS_DESCRIPTION')}</p>
            </div>
            {settingsQuery.isPending ? <CircularProgress size={20} /> : null}
            {isLoadFailed ? (
                <Alert
                    severity="error"
                    action={
                        <Button
                            color="inherit"
                            size="small"
                            onClick={() => void settingsQuery.refetch()}
                        >
                            {t('TRY_AGAIN')}
                        </Button>
                    }
                >
                    {resolveVacuumFailureMessage(
                        settingsQuery.error,
                        t('VACUUM_CLOUD_SETTINGS_LOAD_ERROR')
                    )}
                </Alert>
            ) : null}
            {isFormVisible ? (
                <VacuumCloudSettingsForm
                    savedSettings={toSavedVacuumCloudSettings(settingsQuery.data)}
                />
            ) : null}
        </section>
    );
};

export default VacuumCloudSettingsSection;
