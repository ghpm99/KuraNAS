import { ToggleButton, ToggleButtonGroup } from '@mui/material';
import { useId } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from './VacuumChoicePreference.module.css';

export type VacuumChoiceOption = {
    value: string;
    label: string;
};

type VacuumChoicePreferenceProps = {
    label: string;
    options: VacuumChoiceOption[];
    currentValue: string;
    unlistedValueLabel: string;
    isDisabled: boolean;
    onSelect: (selectedValue: string) => void;
};

const VacuumChoicePreference = ({
    label,
    options,
    currentValue,
    unlistedValueLabel,
    isDisabled,
    onSelect,
}: VacuumChoicePreferenceProps) => {
    const { t } = useI18n();
    const labelId = useId();
    const isCurrentValueOffered = options.some((option) => option.value === currentValue);

    return (
        <div className={styles.preference}>
            <span id={labelId} className={styles.label}>
                {label}
            </span>
            <ToggleButtonGroup
                className={styles.options}
                exclusive
                size="small"
                aria-labelledby={labelId}
                value={isCurrentValueOffered ? currentValue : null}
                disabled={isDisabled}
                onChange={(_event, selectedValue: string | null) => {
                    if (selectedValue === null) return;
                    onSelect(selectedValue);
                }}
            >
                {options.map((option) => (
                    <ToggleButton key={option.value} className={styles.option} value={option.value}>
                        {option.label}
                    </ToggleButton>
                ))}
            </ToggleButtonGroup>
            {!isCurrentValueOffered && (
                <span className={styles.unlistedValue}>
                    {t('VACUUM_PREFERENCE_CURRENT_READ_ONLY', { value: unlistedValueLabel })}
                </span>
            )}
        </div>
    );
};

export default VacuumChoicePreference;
