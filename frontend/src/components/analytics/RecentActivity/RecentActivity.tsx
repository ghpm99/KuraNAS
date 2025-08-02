import { Clock, Eye } from 'lucide-react';

import Card from '../../ui/Card/Card';
import styles from './RecentActivity.module.css';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';

export default function RecentActivity() {
	const { analyticsData } = useAnalytics();
	const { recentFiles, accessedFiles } = analyticsData.recentActivity;

	return (
		<div className={styles.container}>
			<Card title='Arquivos Recentemente Adicionados'>
				<div className={styles.list}>
					{recentFiles.map((file, index) => (
						<div key={index} className={styles.listItem}>
							<div className={styles.itemIcon}>
								<Clock className={styles.icon} />
							</div>
							<div className={styles.itemContent}>
								<div className={styles.itemName}>{file.name}</div>
								<div className={styles.itemMeta}>
									{file.size} • {file.date}
								</div>
							</div>
						</div>
					))}
				</div>
			</Card>

			<Card title='Arquivos Mais Acessados'>
				<div className={styles.list}>
					{accessedFiles.map((file, index) => (
						<div key={index} className={styles.listItem}>
							<div className={styles.itemIcon}>
								<Eye className={styles.icon} />
							</div>
							<div className={styles.itemContent}>
								<div className={styles.itemName}>{file.name}</div>
								<div className={styles.itemMeta}>
									{file.accessCount} acessos • {file.lastAccess}
								</div>
							</div>
						</div>
					))}
				</div>
			</Card>
		</div>
	);
}
