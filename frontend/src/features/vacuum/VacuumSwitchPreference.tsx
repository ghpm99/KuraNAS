import { FormControlLabel, Switch } from '@mui/material';
import styles from './VacuumSwitchPreference.module.css';

type VacuumSwitchPreferenceProps = {
    label: string;
    isChecked: boolean;
    isDisabled: boolean;
    onToggle: (isChecked: boolean) => void;
};

const VacuumSwitchPreference = ({
    label,
    isChecked,
    isDisabled,
    onToggle,
}: VacuumSwitchPreferenceProps) => (
    <FormControlLabel
        className={styles.preference}
        label={label}
        disabled={isDisabled}
        control={
            <Switch
                checked={isChecked}
                onChange={(_event, nextIsChecked) => onToggle(nextIsChecked)}
            />
        }
    />
);

export default VacuumSwitchPreference;
