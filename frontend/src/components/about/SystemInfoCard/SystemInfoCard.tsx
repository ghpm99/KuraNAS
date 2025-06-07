import Card from '@/components/ui/Card/Card';
import styles from './SystemInfoCard.module.css';
import { useAbout } from '@/components/hooks/AboutProvider/AboutContext';
import useI18n from '@/components/i18n/provider/i18nContext';

const SystemInfoCard = () => {
	const { version, platform, lang } = useAbout();
	const { t } = useI18n();
	const systemInfo = {
		projectName: 'Kuranas',
		buildVersion: '20231001',
		buildDate: new Date().toLocaleDateString('pt-BR', {
			year: 'numeric',
			month: '2-digit',
			day: '2-digit',
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit',
		}),
	};

	return (
		<Card title={t('SYSTEM_INFO_TITLE')}>
			<div className={styles.infoGrid}>
				<div className={styles.infoItem}>
					<span className={styles.label}>{t('PROJECT_NAME')}</span>
					<span className={styles.value}>{systemInfo.projectName}</span>
				</div>
				<div className={styles.infoItem}>
					<span className={styles.label}>{t('PROGRAM_VERSION')}</span>
					<span className={styles.value}>{version}</span>
				</div>
				<div className={styles.infoItem}>
					<span className={styles.label}>{t('BUILD_VERSION')}</span>
					<span className={styles.value}>{systemInfo.buildVersion}</span>
				</div>
				<div className={styles.infoItem}>
					<span className={styles.label}>{t('PLATFORM')}</span>
					<span className={`${styles.value} ${styles.platform}`}>
						<span className={styles.platformIcon}>{platform === 'windows' ? '🪟' : '🐧'}</span>
						{platform}
					</span>
				</div>
				<div className={styles.infoItem}>
					<span className={styles.label}>{t('LANGUAGE')}</span>
					<span className={styles.value}>{lang}</span>
				</div>
				<div className={styles.infoItem}>
					<span className={styles.label}>{t('BUILD_DATE')}</span>
					<span className={styles.value}>{systemInfo.buildDate}</span>
				</div>
			</div>
		</Card>
	);
};

export default SystemInfoCard;
