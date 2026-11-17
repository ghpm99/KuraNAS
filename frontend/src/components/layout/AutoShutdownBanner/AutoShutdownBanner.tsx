import { Button } from '@mui/material';
import { Power } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from './AutoShutdownBanner.module.css';
import { useAutoShutdownBanner } from './useAutoShutdownBanner';

const AutoShutdownBanner = () => {
    const { t } = useI18n();
    const { isWarning, secondsLeft, isCancelling, cancelErrorMessage, cancelToday } =
        useAutoShutdownBanner();

    if (!isWarning) return null;

    return (
        <div className={styles.banner} role="alert">
            <Power size={16} aria-hidden="true" />
            <span className={styles.message}>
                {t('AUTO_SHUTDOWN_BANNER_MESSAGE', { seconds: String(secondsLeft) })}
                {cancelErrorMessage ? (
                    <span className={styles.error}> {cancelErrorMessage}</span>
                ) : null}
            </span>
            <Button
                size="small"
                color="inherit"
                variant="outlined"
                disabled={isCancelling}
                onClick={cancelToday}
            >
                {t('AUTO_SHUTDOWN_BANNER_CANCEL')}
            </Button>
        </div>
    );
};

export default AutoShutdownBanner;
