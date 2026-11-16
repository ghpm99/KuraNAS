import { Button } from '@mui/material';
import { ArrowDown, ArrowLeft, ArrowRight, ArrowUp } from 'lucide-react';
import type { KeyboardEvent, ReactNode } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import VacuumBlock from './VacuumBlock';
import styles from './VacuumRemoteControlBlock.module.css';
import { useVacuumRemoteControl } from './useVacuumRemoteControl';
import { directionForArrowKey, type VacuumMovementDirection } from './vacuumRemoteControlKeys';

type DirectionButton = {
    direction: VacuumMovementDirection;
    labelKey: string;
    icon: ReactNode;
};

const directionButtons: DirectionButton[] = [
    {
        direction: 'forward',
        labelKey: 'VACUUM_REMOTE_FORWARD',
        icon: <ArrowUp size={24} />,
    },
    {
        direction: 'left',
        labelKey: 'VACUUM_REMOTE_LEFT',
        icon: <ArrowLeft size={24} />,
    },
    {
        direction: 'right',
        labelKey: 'VACUUM_REMOTE_RIGHT',
        icon: <ArrowRight size={24} />,
    },
    {
        direction: 'backward',
        labelKey: 'VACUUM_REMOTE_BACKWARD',
        icon: <ArrowDown size={24} />,
    },
];

const VacuumRemoteControlBlock = () => {
    const { t } = useI18n();
    const { movingDirection, startMoving, stopMoving } = useVacuumRemoteControl();

    const handleKeyDown = (event: KeyboardEvent) => {
        const direction = directionForArrowKey(event.key);
        if (direction === undefined) return;
        event.preventDefault();
        if (event.repeat) return;
        startMoving(direction);
    };

    const handleKeyUp = (event: KeyboardEvent) => {
        const direction = directionForArrowKey(event.key);
        if (direction === undefined) return;
        stopMoving(direction);
    };

    return (
        <VacuumBlock title={t('VACUUM_REMOTE_TITLE')}>
            <p className={styles.hint}>{t('VACUUM_REMOTE_HINT')}</p>
            <div
                className={styles.pad}
                role="group"
                aria-label={t('VACUUM_REMOTE_TITLE')}
                onKeyDown={handleKeyDown}
                onKeyUp={handleKeyUp}
            >
                {directionButtons.map(({ direction, labelKey, icon }) => (
                    <Button
                        key={direction}
                        className={[styles.direction, styles[direction]].join(' ')}
                        variant={movingDirection === direction ? 'contained' : 'outlined'}
                        aria-label={t(labelKey)}
                        aria-pressed={movingDirection === direction}
                        onPointerDown={() => startMoving(direction)}
                        onPointerUp={() => stopMoving(direction)}
                        onPointerLeave={() => stopMoving(direction)}
                        onPointerCancel={() => stopMoving(direction)}
                        onContextMenu={(event) => event.preventDefault()}
                    >
                        {icon}
                    </Button>
                ))}
            </div>
        </VacuumBlock>
    );
};

export default VacuumRemoteControlBlock;
