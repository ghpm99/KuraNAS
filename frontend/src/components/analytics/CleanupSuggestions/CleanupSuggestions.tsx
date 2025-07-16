import { AlertTriangle, FileX, Trash2 } from 'lucide-react';

import Card from '../../ui/Card/Card';
import styles from './CleanupSuggestions.module.css';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';

export default function CleanupSuggestions() {
	const { analyticsData } = useAnalytics();
	const { cleanup } = analyticsData;

	return (
		<div className={styles.section}>
			{cleanup.criticalSpace && (
				<div className={styles.alert}>
					<AlertTriangle className={styles.alertIcon} />
					<div className={styles.alertContent}>
						<div className={styles.alertTitle}>Espaço Crítico</div>
						<div className={styles.alertMessage}>
							O uso de disco ultrapassou 90%. Considere limpar arquivos desnecessários.
						</div>
					</div>
				</div>
			)}

			<div className={styles.cardsGrid}>
				<Card title='Arquivos Grandes Não Acessados'>
					<div className={styles.list}>
						{cleanup.oldLargeFiles.map((file, index) => (
							<div key={index} className={styles.listItem}>
								<div className={styles.itemIcon}>
									<FileX className={styles.icon} />
								</div>
								<div className={styles.itemContent}>
									<div className={styles.itemName}>{file.name}</div>
									<div className={styles.itemMeta}>
										{file.size} • {file.path}
									</div>
								</div>
							</div>
						))}
					</div>
				</Card>

				<Card title='Arquivos com Nomes Similares'>
					<div className={styles.list}>
						{cleanup.similarNames.map((similar, index) => (
							<div key={index} className={styles.listItem}>
								<div className={styles.itemIcon}>
									<Trash2 className={styles.icon} />
								</div>
								<div className={styles.itemContent}>
									<div className={styles.itemName}>{similar.name1}</div>
									<div className={styles.itemName}>{similar.name2}</div>
									<div className={styles.itemMeta}>Similaridade: {similar.similarity}%</div>
								</div>
							</div>
						))}
					</div>
				</Card>
			</div>
		</div>
	);
}
