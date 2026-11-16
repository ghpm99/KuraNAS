import { Alert } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import { formatDateTime } from '@/shared/utils/formatDateTime';
import { formatElapsedDuration } from './formatElapsedDuration';
import VacuumBlock from './VacuumBlock';
import VacuumFaultCodes from './VacuumFaultCodes';
import styles from './VacuumStatusBlock.module.css';
import { isCleaningSessionState, toVacuumState, vacuumStateLabelKeys } from './vacuumStateCatalog';
import { useVacuumStatus } from './useVacuumStatus';

const VacuumStatusBlock = () => {
    const { t } = useI18n();
    const statusQuery = useVacuumStatus();
    const status = statusQuery.data;
    const state = toVacuumState(status?.state);
    const isStale = statusQuery.isError && status !== undefined;

    return (
        <VacuumBlock
            title={t('VACUUM_STATUS_TITLE')}
            isLoading={statusQuery.isPending}
            hasLoadError={statusQuery.isError && status === undefined}
            loadError={statusQuery.error}
            onRetry={() => void statusQuery.refetch()}
        >
            <div className={styles.summary}>
                <p className={styles.stateName}>{t(vacuumStateLabelKeys[state])}</p>
                <ul className={styles.facts}>
                    {typeof status?.battery_percent === 'number' && (
                        <li>
                            {t('VACUUM_STATUS_BATTERY', {
                                percent: String(status.battery_percent),
                            })}
                        </li>
                    )}
                    {isCleaningSessionState(state) && (
                        <>
                            <li>
                                {t('VACUUM_STATUS_ELAPSED', {
                                    duration: formatElapsedDuration(status?.cleaning_seconds),
                                })}
                            </li>
                            <li>
                                {t('VACUUM_STATUS_AREA', {
                                    area: String(status?.cleaning_area ?? 0),
                                })}
                            </li>
                        </>
                    )}
                </ul>
                <VacuumFaultCodes faultCodes={status?.fault_codes} />
                {isStale && (
                    <Alert severity="warning" role="status">
                        {t('VACUUM_STATUS_STALE', { time: formatDateTime(status?.read_at) })}
                    </Alert>
                )}
            </div>
        </VacuumBlock>
    );
};

export default VacuumStatusBlock;
