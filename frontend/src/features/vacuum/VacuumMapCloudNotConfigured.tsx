import { Alert, Button } from '@mui/material';
import { Link as RouterLink } from 'react-router-dom';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getSettingsSectionRoute } from '@/components/settings/settingsSections';
import styles from './VacuumMapCloudNotConfigured.module.css';

const VacuumMapCloudNotConfigured = () => {
    const { t } = useI18n();

    return (
        <Alert severity="info" role="status" className={styles.notice}>
            <p className={styles.explanation}>{t('VACUUM_MAP_CLOUD_NOT_CONFIGURED_DESCRIPTION')}</p>
            <Button
                className={styles.settingsLink}
                variant="contained"
                component={RouterLink}
                to={getSettingsSectionRoute('vacuum-cloud')}
            >
                {t('VACUUM_MAP_CLOUD_NOT_CONFIGURED_ACTION')}
            </Button>
        </Alert>
    );
};

export default VacuumMapCloudNotConfigured;
