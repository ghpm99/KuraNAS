import useI18n from '@/components/i18n/provider/i18nContext';
import { useImageDomainHeader } from '@/components/images/useImageDomainHeader';
import styles from './ImageDomainHeader.module.css';

const ImageDomainHeader = () => {
	const { t } = useI18n();
	const { titleKey, descriptionKey } = useImageDomainHeader();

	return (
		<header className={styles.header}>
			<span className={styles.eyebrow}>{t('NAV_IMAGES')}</span>
			<h1 className={styles.title}>{t(titleKey)}</h1>
			<p className={styles.description}>{t(descriptionKey)}</p>
		</header>
	);
};

export default ImageDomainHeader;
