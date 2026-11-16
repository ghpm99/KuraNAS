import { Alert, Button } from '@mui/material';
import { Link as RouterLink } from 'react-router-dom';
import { getSettingsSectionRoute } from '@/components/settings/settingsSections';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from './VacuumNotConfigured.module.css';

type VacuumNotConfiguredProps = {
    isDeviceIdMissing: boolean;
    isLocalKeyMissing: boolean;
    serverMessage?: string;
};

const VacuumNotConfigured = ({
    isDeviceIdMissing,
    isLocalKeyMissing,
    serverMessage,
}: VacuumNotConfiguredProps) => {
    const { t } = useI18n();

    return (
        <Alert severity="info" role="status" className={styles.notice}>
            <p className={styles.explanation}>{t('VACUUM_NOT_CONFIGURED_DESCRIPTION')}</p>
            <ul className={styles.missing}>
                {isDeviceIdMissing && <li>{t('VACUUM_NOT_CONFIGURED_DEVICE_ID')}</li>}
                {isLocalKeyMissing && <li>{t('VACUUM_NOT_CONFIGURED_LOCAL_KEY')}</li>}
            </ul>
            {serverMessage && <p className={styles.explanation}>{serverMessage}</p>}
            <Button
                className={styles.settingsLink}
                variant="contained"
                component={RouterLink}
                to={getSettingsSectionRoute('vacuum')}
            >
                {t('VACUUM_NOT_CONFIGURED_ACTION')}
            </Button>
        </Alert>
    );
};

export default VacuumNotConfigured;
