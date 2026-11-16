import useI18n from '@/components/i18n/provider/i18nContext';
import VacuumBlock from './VacuumBlock';
import styles from './VacuumTotalsBlock.module.css';
import { formatSecondsAsHours } from './formatSecondsAsHours';
import { useVacuumTotals } from './useVacuumTotals';

type TotalEntry = {
    labelKey: string;
    amount: string;
};

const VacuumTotalsBlock = () => {
    const { t } = useI18n();
    const totalsQuery = useVacuumTotals();
    const totals = totalsQuery.data;

    const entries: TotalEntry[] = [];
    if (typeof totals?.cleaning_seconds === 'number') {
        entries.push({
            labelKey: 'VACUUM_TOTALS_TIME',
            amount: t('VACUUM_TOTALS_HOURS', { hours: formatSecondsAsHours(totals.cleaning_seconds) }),
        });
    }
    if (typeof totals?.cleaning_area === 'number') {
        entries.push({
            labelKey: 'VACUUM_TOTALS_AREA',
            amount: t('VACUUM_TOTALS_SQUARE_METERS', { area: String(totals.cleaning_area) }),
        });
    }
    if (typeof totals?.cleaning_count === 'number') {
        entries.push({ labelKey: 'VACUUM_TOTALS_COUNT', amount: String(totals.cleaning_count) });
    }

    return (
        <VacuumBlock
            title={t('VACUUM_TOTALS_TITLE')}
            isLoading={totalsQuery.isPending}
            hasLoadError={totalsQuery.isError && totals === undefined}
            loadError={totalsQuery.error}
            onRetry={() => void totalsQuery.refetch()}
        >
            <dl className={styles.totals}>
                {entries.map((entry) => (
                    <div key={entry.labelKey} className={styles.total}>
                        <dt className={styles.totalLabel}>{t(entry.labelKey)}</dt>
                        <dd className={styles.totalAmount}>{entry.amount}</dd>
                    </div>
                ))}
            </dl>
        </VacuumBlock>
    );
};

export default VacuumTotalsBlock;
