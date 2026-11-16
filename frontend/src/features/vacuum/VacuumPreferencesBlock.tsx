import useI18n from '@/components/i18n/provider/i18nContext';
import type { VacuumPreferences } from '@/types/vacuum';
import VacuumBlock from './VacuumBlock';
import VacuumChoicePreference, { type VacuumChoiceOption } from './VacuumChoicePreference';
import VacuumSwitchPreference from './VacuumSwitchPreference';
import VacuumVolumePreference from './VacuumVolumePreference';
import styles from './VacuumPreferencesBlock.module.css';
import { useVacuumPreferences } from './useVacuumPreferences';

type VacuumSwitchField =
    | 'is_carpet_boost'
    | 'is_resume_after_charge'
    | 'is_do_not_disturb'
    | 'is_auto_empty';

const suctionOptions: VacuumChoiceOption[] = [
    { value: 'quiet', label: 'VACUUM_SUCTION_QUIET' },
    { value: 'auto', label: 'VACUUM_SUCTION_AUTO' },
    { value: 'strong', label: 'VACUUM_SUCTION_STRONG' },
    { value: 'max', label: 'VACUUM_SUCTION_MAX' },
];

const waterLevelOptions: VacuumChoiceOption[] = [
    { value: 'low', label: 'VACUUM_WATER_LOW' },
    { value: 'mid', label: 'VACUUM_WATER_MID' },
    { value: 'high', label: 'VACUUM_WATER_HIGH' },
];

const autoEmptyFrequencyOptions: VacuumChoiceOption[] = ['0', '1', '2', '3'].map((frequency) => ({
    value: frequency,
    label: frequency,
}));

const unlistedValueLabelKeys: Record<string, string> = {
    mop: 'VACUUM_SUCTION_MOP',
};

const isReported = <Reported,>(preference: Reported | null | undefined): preference is Reported =>
    preference !== null && preference !== undefined;

const hasAnyReportedPreference = (preferences: Partial<VacuumPreferences>): boolean =>
    Object.values(preferences).some(isReported);

const VacuumPreferencesBlock = () => {
    const { t } = useI18n();
    const { preferencesQuery, pendingPreferenceName, savePreference } = useVacuumPreferences();
    const preferences: Partial<VacuumPreferences> = preferencesQuery.data ?? {};

    const translateOptions = (options: VacuumChoiceOption[]): VacuumChoiceOption[] =>
        options.map((option) => ({ ...option, label: t(option.label) }));

    const unlistedValueLabelFor = (reportedValue: string): string =>
        t(unlistedValueLabelKeys[reportedValue] ?? 'VACUUM_VALUE_UNKNOWN');

    const renderSwitchPreference = (field: VacuumSwitchField, labelKey: string) => {
        const reportedValue = preferences[field];
        if (!isReported(reportedValue)) return null;
        return (
            <VacuumSwitchPreference
                label={t(labelKey)}
                isChecked={reportedValue}
                isDisabled={pendingPreferenceName === field}
                onToggle={(isChecked) => void savePreference({ [field]: isChecked })}
            />
        );
    };

    return (
        <VacuumBlock
            title={t('VACUUM_PREFERENCES_TITLE')}
            isLoading={preferencesQuery.isPending}
            hasLoadError={preferencesQuery.isError && preferencesQuery.data === undefined}
            loadError={preferencesQuery.error}
            onRetry={() => void preferencesQuery.refetch()}
        >
            {!hasAnyReportedPreference(preferences) && (
                <p className={styles.empty}>{t('VACUUM_PREFERENCES_EMPTY')}</p>
            )}
            <div className={styles.preferences}>
                {isReported(preferences.suction) && (
                    <VacuumChoicePreference
                        label={t('VACUUM_PREFERENCE_SUCTION')}
                        options={translateOptions(suctionOptions)}
                        currentValue={preferences.suction}
                        unlistedValueLabel={unlistedValueLabelFor(preferences.suction)}
                        isDisabled={pendingPreferenceName === 'suction'}
                        onSelect={(suction) => void savePreference({ suction })}
                    />
                )}
                {isReported(preferences.water_level) && (
                    <VacuumChoicePreference
                        label={t('VACUUM_PREFERENCE_WATER_LEVEL')}
                        options={translateOptions(waterLevelOptions)}
                        currentValue={preferences.water_level}
                        unlistedValueLabel={unlistedValueLabelFor(preferences.water_level)}
                        isDisabled={pendingPreferenceName === 'water_level'}
                        onSelect={(waterLevel) => void savePreference({ water_level: waterLevel })}
                    />
                )}
                {renderSwitchPreference('is_carpet_boost', 'VACUUM_PREFERENCE_CARPET_BOOST')}
                {renderSwitchPreference(
                    'is_resume_after_charge',
                    'VACUUM_PREFERENCE_RESUME_AFTER_CHARGE'
                )}
                {isReported(preferences.volume) && (
                    <VacuumVolumePreference
                        label={t('VACUUM_PREFERENCE_VOLUME')}
                        volume={preferences.volume}
                        isDisabled={pendingPreferenceName === 'volume'}
                        onCommit={(volume) => savePreference({ volume })}
                    />
                )}
                {renderSwitchPreference('is_do_not_disturb', 'VACUUM_PREFERENCE_DO_NOT_DISTURB')}
                {renderSwitchPreference('is_auto_empty', 'VACUUM_PREFERENCE_AUTO_EMPTY')}
                {isReported(preferences.auto_empty_frequency) && (
                    <VacuumChoicePreference
                        label={t('VACUUM_PREFERENCE_AUTO_EMPTY_FREQUENCY')}
                        options={autoEmptyFrequencyOptions}
                        currentValue={preferences.auto_empty_frequency}
                        unlistedValueLabel={unlistedValueLabelFor(preferences.auto_empty_frequency)}
                        isDisabled={pendingPreferenceName === 'auto_empty_frequency'}
                        onSelect={(frequency) =>
                            void savePreference({ auto_empty_frequency: frequency })
                        }
                    />
                )}
            </div>
        </VacuumBlock>
    );
};

export default VacuumPreferencesBlock;
