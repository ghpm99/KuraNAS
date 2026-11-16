import { Button } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from './VacuumCommandsBlock.module.css';
import VacuumBlock from './VacuumBlock';
import {
    isVacuumCommandAvailable,
    vacuumCommandLabelKeys,
    vacuumCommandsInDisplayOrder,
} from './vacuumCommandAvailability';
import { useVacuumCommands } from './useVacuumCommands';

const VacuumCommandsBlock = () => {
    const { t } = useI18n();
    const { state, pendingCommand, sendCommand } = useVacuumCommands();

    return (
        <VacuumBlock title={t('VACUUM_COMMANDS_TITLE')}>
            <div className={styles.commands}>
                {vacuumCommandsInDisplayOrder.map((command) => (
                    <Button
                        key={command}
                        className={styles.command}
                        variant="outlined"
                        disabled={!isVacuumCommandAvailable(command, state) || pendingCommand === command}
                        onClick={() => sendCommand(command)}
                    >
                        {t(vacuumCommandLabelKeys[command])}
                    </Button>
                ))}
            </div>
        </VacuumBlock>
    );
};

export default VacuumCommandsBlock;
