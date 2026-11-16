import { Slider } from '@mui/material';
import { useId, useState } from 'react';
import styles from './VacuumVolumePreference.module.css';

const minimumVolume = 0;
const maximumVolume = 10;

type VacuumVolumePreferenceProps = {
    label: string;
    volume: number;
    isDisabled: boolean;
    onCommit: (committedVolume: number) => Promise<void>;
};

const VacuumVolumePreference = ({
    label,
    volume,
    isDisabled,
    onCommit,
}: VacuumVolumePreferenceProps) => {
    const labelId = useId();
    const [draggedVolume, setDraggedVolume] = useState<number | null>(null);

    const commitVolume = async (committedVolume: number) => {
        try {
            await onCommit(committedVolume);
        } finally {
            setDraggedVolume(null);
        }
    };

    return (
        <div className={styles.preference}>
            <span id={labelId} className={styles.label}>
                {label}
            </span>
            <Slider
                className={styles.slider}
                aria-labelledby={labelId}
                min={minimumVolume}
                max={maximumVolume}
                step={1}
                marks
                valueLabelDisplay="auto"
                disabled={isDisabled}
                value={draggedVolume ?? volume}
                onChange={(_event, nextVolume) => setDraggedVolume(nextVolume)}
                onChangeCommitted={(_event, committedVolume) => void commitVolume(committedVolume)}
            />
        </div>
    );
};

export default VacuumVolumePreference;
