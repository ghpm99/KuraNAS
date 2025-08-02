import { Database, File, Folder, HardDrive } from 'lucide-react';

import styles from './StorageOverviewCards.module.css';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';

export default function StorageOverviewCards() {
	const { analyticsData } = useAnalytics();
	const { storageOverview } = analyticsData;

	return (
		<div className={styles.grid}>
			<div className={styles.card}>
				<div className={styles.cardContent}>
					<div className={styles.iconContainer}>
						<HardDrive className={styles.icon} />
					</div>
					<div className={styles.info}>
						<div className={styles.label}>Espaço Utilizado</div>
						<div className={styles.value}>{storageOverview.totalUsedSpace}</div>
					</div>
				</div>
			</div>

			<div className={styles.card}>
				<div className={styles.cardContent}>
					<div className={styles.iconContainer}>
						<File className={styles.icon} />
					</div>
					<div className={styles.info}>
						<div className={styles.label}>Arquivos Armazenados</div>
						<div className={styles.value}>{storageOverview.totalFiles.toLocaleString()}</div>
					</div>
				</div>
			</div>

			<div className={styles.card}>
				<div className={styles.cardContent}>
					<div className={styles.iconContainer}>
						<Folder className={styles.icon} />
					</div>
					<div className={styles.info}>
						<div className={styles.label}>Pastas Armazenadas</div>
						<div className={styles.value}>{storageOverview.totalFolders.toLocaleString()}</div>
					</div>
				</div>
			</div>

			<div className={styles.card}>
				<div className={styles.cardContent}>
					<div className={styles.iconContainer}>
						<Database className={styles.icon} />
					</div>
					<div className={styles.info}>
						<div className={styles.label}>Espaço Livre</div>
						<div className={styles.value}>{storageOverview.availableSpace}</div>
					</div>
				</div>
			</div>
		</div>
	);
}
