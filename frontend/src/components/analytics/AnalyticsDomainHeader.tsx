import useI18n from '@/components/i18n/provider/i18nContext';
import { useAnalyticsDomainHeader } from '@/components/analytics/useAnalyticsDomainHeader';
import styles from './AnalyticsDomainHeader.module.css';

const AnalyticsDomainHeader = () => {
	const { t } = useI18n();
	const { titleKey, descriptionKey } = useAnalyticsDomainHeader();

	return (
		<header className={styles.header}>
			<span className={styles.eyebrow}>{t('ANALYTICS')}</span>
			<h1 className={styles.title}>{t(titleKey)}</h1>
			<p className={styles.description}>{t(descriptionKey)}</p>
		</header>
	);
};

export default AnalyticsDomainHeader;
