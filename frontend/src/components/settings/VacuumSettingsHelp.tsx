import { Alert } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from './VacuumSettingsSection.module.css';

const VacuumSettingsHelp = () => {
    const { t } = useI18n();

    return (
        <Alert severity="info" className={styles.help}>
            <p className={styles.helpText}>{t('VACUUM_SETTINGS_HELP_WHERE')}</p>
            <p className={styles.helpText}>{t('VACUUM_SETTINGS_HELP_KEY_CHANGES')}</p>
        </Alert>
    );
};

export default VacuumSettingsHelp;
