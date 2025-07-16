import { FolderX } from 'lucide-react';

import Card from '../../ui/Card/Card';
import styles from './EmptyFoldersSection.module.css';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';

export default function EmptyFoldersSection() {
	const { analyticsData } = useAnalytics();
	const { organization } = analyticsData;

	return (
		<div className={styles.section}>
			<div className={styles.card}>
				<div className={styles.cardContent}>
					<div className={styles.iconContainer}>
						<FolderX className={styles.icon} />
					</div>
					<div className={styles.info}>
						<div className={styles.label}>Pastas Vazias</div>
						<div className={styles.value}>{organization.emptyFolders}</div>
					</div>
				</div>
			</div>

			<Card title='Caminhos Vazios'>
				<div className={styles.pathsList}>
					{organization.emptyPaths.map((path, index) => (
						<div key={index} className={styles.pathItem}>
							<FolderX className={styles.pathIcon} />
							<span className={styles.pathText}>{path}</span>
						</div>
					))}
				</div>
			</Card>
		</div>
	);
}
