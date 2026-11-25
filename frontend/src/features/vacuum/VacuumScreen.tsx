import { CircularProgress } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import PageContainer from '@/components/layout/PageContainer';
import PageHeader from '@/components/layout/PageHeader';
import VacuumCommandsBlock from './VacuumCommandsBlock';
import VacuumConsumablesBlock from './VacuumConsumablesBlock';
import VacuumMapBlock from './VacuumMapBlock';
import VacuumNotConfigured from './VacuumNotConfigured';
import VacuumPreferencesBlock from './VacuumPreferencesBlock';
import VacuumRemoteControlBlock from './VacuumRemoteControlBlock';
import styles from './VacuumScreen.module.css';
import VacuumStatusBlock from './VacuumStatusBlock';
import VacuumTotalsBlock from './VacuumTotalsBlock';
import { useVacuumAvailability } from './useVacuumAvailability';

const VacuumScreen = () => {
    const { t } = useI18n();
    const availability = useVacuumAvailability();

    const renderBody = () => {
        if (availability.isChecking) {
            return (
                <div className={styles.checking} role="status">
                    <CircularProgress size={20} />
                    <span>{t('LOADING')}</span>
                </div>
            );
        }
        if (availability.isNotConfigured) {
            return (
                <VacuumNotConfigured
                    isDeviceIdMissing={availability.isDeviceIdMissing}
                    isLocalKeyMissing={availability.isLocalKeyMissing}
                    serverMessage={availability.serverMessage}
                />
            );
        }
        return (
            <div className={styles.blocks}>
                <VacuumStatusBlock />
                <VacuumCommandsBlock />
                <VacuumMapBlock />
                <VacuumRemoteControlBlock />
                <VacuumPreferencesBlock />
                <VacuumConsumablesBlock />
                <VacuumTotalsBlock />
            </div>
        );
    };

    return (
        <PageContainer>
            <PageHeader title={t('VACUUM_PAGE_TITLE')} subtitle={t('VACUUM_PAGE_DESCRIPTION')} />
            {renderBody()}
        </PageContainer>
    );
};

export default VacuumScreen;
