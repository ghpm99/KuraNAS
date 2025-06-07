import Card from '@/components/ui/Card/Card';
import styles from './StatusCard.module.css';
import { useAbout } from '@/components/hooks/AboutProvider/AboutContext';
import useI18n from '@/components/i18n/provider/i18nContext';

const StatusCard = () => {
	const { enable_workers, path, uptime } = useAbout();
	const { t } = useI18n();

	return (
		<Card title={t('STATUS_SYSTEM_TITLE')}>
			<div className={styles.statusGrid}>
				<div className={styles.statusItem}>
					<div className={styles.statusHeader}>
						<span className={styles.label}>{t('WORKERS')}</span>
						<span className={`${styles.status} ${enable_workers ? styles.enabled : styles.disabled}`}>
							{enable_workers ? t('ENABLED_WORKERS') : t('DISABLED_WORKERS')}
						</span>
					</div>
					<div className={styles.statusDescription}>
						{enable_workers ? t('ENABLED_WORKERS_DESCRIPTION') : t('DISABLED_WORKERS_DESCRIPTION')}
					</div>
				</div>

				<div className={styles.statusItem}>
					<div className={styles.statusHeader}>
						<span className={styles.label}>{t('WATCH_PATH')}</span>
					</div>
					<div className={styles.folderPath}>{path}</div>
					<div className={styles.statusDescription}>{t('WATCH_PATH_DESCRIPTION')}</div>
				</div>

				<div className={styles.statusItem}>
					<div className={styles.statusHeader}>
						<span className={styles.label}>{t('UPTIME')}</span>
					</div>
					<div className={styles.uptime}>{uptime}</div>
					<div className={styles.statusDescription}>{t('UPTIME_DESCRIPTION')}</div>
				</div>
			</div>
		</Card>
	);
};

export default StatusCard;
