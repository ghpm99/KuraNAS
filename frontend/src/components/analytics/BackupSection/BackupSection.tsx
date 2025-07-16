import { Calendar, CheckCircle, HardDrive, XCircle } from 'lucide-react';

import Card from '../../ui/Card/Card';
import styles from './BackupSection.module.css';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';

export default function BackupSection() {
	const { analyticsData } = useAnalytics();
	const { backup } = analyticsData;

	const getStatusIcon = (status: string) => {
		switch (status) {
			case 'success':
				return <CheckCircle className={`${styles.statusIcon} ${styles.success}`} />;
			case 'failed':
				return <XCircle className={`${styles.statusIcon} ${styles.failed}`} />;
			default:
				return <Calendar className={`${styles.statusIcon} ${styles.pending}`} />;
		}
	};

	const getStatusText = (status: string) => {
		switch (status) {
			case 'success':
				return 'Sucesso';
			case 'failed':
				return 'Falhou';
			default:
				return 'Pendente';
		}
	};

	return (
		<div className={styles.section}>
			<div className={styles.cardsGrid}>
				<div className={styles.card}>
					<div className={styles.cardContent}>
						<div className={styles.iconContainer}>
							<Calendar className={styles.icon} />
						</div>
						<div className={styles.info}>
							<div className={styles.label}>Último Backup</div>
							<div className={styles.value}>{backup.lastBackup}</div>
						</div>
					</div>
				</div>

				<div className={styles.card}>
					<div className={styles.cardContent}>
						<div className={styles.iconContainer}>
							<HardDrive className={styles.icon} />
						</div>
						<div className={styles.info}>
							<div className={styles.label}>Tamanho do Último Backup</div>
							<div className={styles.value}>{backup.lastBackupSize}</div>
						</div>
					</div>
				</div>
			</div>

			<Card title='Histórico de Backups'>
				<div className={styles.tableContainer}>
					<table className={styles.table}>
						<thead>
							<tr>
								<th>Data</th>
								<th>Tamanho</th>
								<th>Status</th>
							</tr>
						</thead>
						<tbody>
							{backup.history.map((item, index) => (
								<tr key={index}>
									<td className={styles.dateCell}>{item.date}</td>
									<td className={styles.sizeCell}>{item.size}</td>
									<td className={styles.statusCell}>
										<div className={styles.statusContainer}>
											{getStatusIcon(item.status)}
											<span>{getStatusText(item.status)}</span>
										</div>
									</td>
								</tr>
							))}
						</tbody>
					</table>
				</div>
			</Card>
		</div>
	);
}
