import { Button } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import VacuumBlock from './VacuumBlock';
import VacuumResetConfirmDialog from './VacuumResetConfirmDialog';
import styles from './VacuumConsumablesBlock.module.css';
import { isVacuumConsumablePart, vacuumPartLabelKeys } from './vacuumConsumableParts';
import { formatSecondsAsHours } from './formatSecondsAsHours';
import { useVacuumConsumables } from './useVacuumConsumables';

const VacuumConsumablesBlock = () => {
    const { t } = useI18n();
    const {
        consumablesQuery,
        partAwaitingConfirmation,
        isResetting,
        askToReset,
        cancelReset,
        confirmReset,
    } = useVacuumConsumables();
    const consumables = Array.isArray(consumablesQuery.data?.items)
        ? consumablesQuery.data.items
        : [];

    const partLabelFor = (part: string): string =>
        t(isVacuumConsumablePart(part) ? vacuumPartLabelKeys[part] : 'VACUUM_PART_UNKNOWN');

    const renderResetButton = (part: string) => {
        if (!isVacuumConsumablePart(part)) return null;
        return (
            <Button
                className={styles.reset}
                variant="outlined"
                aria-label={t('VACUUM_RESET_PART_LABEL', { part: partLabelFor(part) })}
                onClick={() => askToReset(part)}
            >
                {t('VACUUM_RESET_ACTION')}
            </Button>
        );
    };

    return (
        <VacuumBlock
            title={t('VACUUM_CONSUMABLES_TITLE')}
            isLoading={consumablesQuery.isPending}
            hasLoadError={consumablesQuery.isError && consumablesQuery.data === undefined}
            loadError={consumablesQuery.error}
            onRetry={() => void consumablesQuery.refetch()}
        >
            {consumables.length === 0 && (
                <p className={styles.empty}>{t('VACUUM_CONSUMABLES_EMPTY')}</p>
            )}
            <ul className={styles.parts}>
                {consumables.map((consumable) => (
                    <li key={consumable.part} className={styles.part}>
                        <div className={styles.partDescription}>
                            <span className={styles.partName}>{partLabelFor(consumable.part)}</span>
                            <span className={styles.partLife}>
                                {t('VACUUM_CONSUMABLE_HOURS', {
                                    current: formatSecondsAsHours(consumable.seconds),
                                    maximum: formatSecondsAsHours(consumable.max_seconds),
                                })}
                            </span>
                        </div>
                        {renderResetButton(consumable.part)}
                    </li>
                ))}
            </ul>
            <VacuumResetConfirmDialog
                partLabel={
                    partAwaitingConfirmation === null
                        ? undefined
                        : partLabelFor(partAwaitingConfirmation)
                }
                isResetting={isResetting}
                onCancel={cancelReset}
                onConfirm={confirmReset}
            />
        </VacuumBlock>
    );
};

export default VacuumConsumablesBlock;
