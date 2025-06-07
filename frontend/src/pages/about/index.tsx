import StatusCard from '@/components/about/StatusCard/StatusCard';
import styles from './about.module.css';
import SystemInfoCard from '@/components/about/SystemInfoCard/SystemInfoCard';
import TechnicalInfoCard from '@/components/about/TechnicalInfoCard/TechnicalInfoCard';
import useI18n from '@/components/i18n/provider/i18nContext';

const AboutPage = () => {
	const { t } = useI18n();
	return (
		<div className={styles.content}>
			<div className={styles.header}>
				<h1 className={styles.pageTitle}>{t('ABOUT_PAGE_TITLE')}</h1>
				<p className={styles.pageDescription}>{t('ABOUT_PAGE_DESCRIPTION')}</p>
			</div>

			<div className={styles.grid}>
				<div className={styles.leftColumn}>
					<SystemInfoCard />
					<TechnicalInfoCard />
				</div>
				<div className={styles.rightColumn}>
					<StatusCard />
				</div>
			</div>
		</div>
	);
};

export default AboutPage;
