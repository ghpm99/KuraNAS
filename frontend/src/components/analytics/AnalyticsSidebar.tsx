import { Link } from 'react-router-dom';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useAnalyticsNavigation } from '@/components/analytics/useAnalyticsNavigation';
import styles from './AnalyticsSidebar.module.css';

const AnalyticsSidebar = () => {
    const { t } = useI18n();
    const { items } = useAnalyticsNavigation();

    return (
        <nav className={styles.nav} aria-label={t('ANALYTICS_NAVIGATION_LABEL')}>
            <p className={styles.label}>{t('ANALYTICS_NAVIGATION_LABEL')}</p>
            <div className={styles.list}>
                {items.map((item) => (
                    <Link
                        key={item.key}
                        to={item.href}
                        className={
                            item.isActive ? `${styles.link} ${styles.linkActive}` : styles.link
                        }
                    >
                        <span className={styles.title}>{t(item.labelKey)}</span>
                        <span className={styles.description}>{t(item.descriptionKey)}</span>
                    </Link>
                ))}
            </div>
        </nav>
    );
};

export default AnalyticsSidebar;
