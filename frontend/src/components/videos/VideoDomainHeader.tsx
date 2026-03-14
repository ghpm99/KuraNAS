import useI18n from '@/components/i18n/provider/i18nContext';
import { useVideoDomainHeader } from '@/components/videos/useVideoDomainHeader';
import styles from './VideoDomainHeader.module.css';

const VideoDomainHeader = () => {
	const { t } = useI18n();
	const { titleKey, descriptionKey } = useVideoDomainHeader();

	return (
		<header className={styles.header}>
			<span className={styles.eyebrow}>{t('NAV_VIDEOS')}</span>
			<h1 className={styles.title}>{t(titleKey)}</h1>
			<p className={styles.description}>{t(descriptionKey)}</p>
		</header>
	);
};

export default VideoDomainHeader;
