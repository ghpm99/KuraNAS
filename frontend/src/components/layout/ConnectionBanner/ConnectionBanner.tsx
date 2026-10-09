import { Button } from '@mui/material';
import { WifiOff } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from './ConnectionBanner.module.css';
import { useServerConnection } from './useServerConnection';

const ConnectionBanner = () => {
    const { t } = useI18n();
    const { isUnavailable, retryNow } = useServerConnection();

    if (!isUnavailable) return null;

    return (
        <div className={styles.banner} role="status" aria-live="polite">
            <WifiOff size={16} aria-hidden="true" />
            <span className={styles.message}>{t('CONNECTION_BANNER_UNAVAILABLE')}</span>
            <Button size="small" color="inherit" variant="outlined" onClick={() => void retryNow()}>
                {t('CONNECTION_RETRY_NOW')}
            </Button>
        </div>
    );
};

export default ConnectionBanner;
