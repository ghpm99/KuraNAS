import { File, Trash2 } from 'lucide-react';

import Card from '../../ui/Card/Card';
import styles from './TrashSection.module.css';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';

export default function TrashSection() {
	const { analyticsData } = useAnalytics();
	const { trash } = analyticsData;

	return (
		<div className={styles.section}>
			<div className={styles.cardsGrid}>
				<div className={styles.card}>
					<div className={styles.cardContent}>
						<div className={styles.iconContainer}>
							<File className={styles.icon} />
						</div>
						<div className={styles.info}>
							<div className={styles.label}>Arquivos na Lixeira</div>
							<div className={styles.value}>{trash.totalFiles}</div>
						</div>
					</div>
				</div>

				<div className={styles.card}>
					<div className={styles.cardContent}>
						<div className={styles.iconContainer}>
							<Trash2 className={styles.icon} />
						</div>
						<div className={styles.info}>
							<div className={styles.label}>Espaço Ocupado</div>
							<div className={styles.value}>{trash.totalSpace}</div>
						</div>
					</div>
				</div>
			</div>

			<Card title='Arquivos na Lixeira'>
				<div className={styles.tableContainer}>
					<table className={styles.table}>
						<thead>
							<tr>
								<th>Nome</th>
								<th>Tamanho</th>
								<th>Data de Exclusão</th>
							</tr>
						</thead>
						<tbody>
							{trash.files.map((file, index) => (
								<tr key={index}>
									<td>
										<div className={styles.fileCell}>
											<Trash2 className={styles.fileIcon} />
											<span className={styles.fileName}>{file.name}</span>
										</div>
									</td>
									<td className={styles.sizeCell}>{file.size}</td>
									<td className={styles.dateCell}>{file.deletedDate}</td>
								</tr>
							))}
						</tbody>
					</table>
				</div>
			</Card>
		</div>
	);
}
