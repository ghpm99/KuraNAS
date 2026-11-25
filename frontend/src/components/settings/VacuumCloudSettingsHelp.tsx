import { Alert } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from './VacuumSettingsSection.module.css';

const VacuumCloudSettingsHelp = () => {
    const { t } = useI18n();

    return (
        <Alert severity="info" className={styles.help}>
            <p className={styles.helpText}>{t('VACUUM_CLOUD_SETTINGS_HELP_WHERE')}</p>
            <p className={styles.helpText}>{t('VACUUM_CLOUD_SETTINGS_HELP_SERVICE')}</p>
            <p className={styles.helpText}>{t('VACUUM_CLOUD_SETTINGS_HELP_REGION')}</p>
            <p className={styles.helpText}>{t('VACUUM_CLOUD_SETTINGS_HELP_MAP_ONLY')}</p>
        </Alert>
    );
};

export default VacuumCloudSettingsHelp;
