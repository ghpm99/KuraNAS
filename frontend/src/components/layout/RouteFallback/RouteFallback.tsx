import { CircularProgress } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from './RouteFallback.module.css';

const RouteFallback = () => {
    const { t } = useI18n();

    return (
        <div className={styles.fallback} role="status" aria-busy="true" aria-live="polite">
            <CircularProgress size={28} aria-label={t('ROUTE_LOADING')} />
            <span className={styles.label}>{t('ROUTE_LOADING')}</span>
        </div>
    );
};

export default RouteFallback;
