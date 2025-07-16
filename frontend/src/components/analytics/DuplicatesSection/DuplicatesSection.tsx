import { Copy, HardDrive } from 'lucide-react';

import Card from '../../ui/Card/Card';
import styles from './DuplicatesSection.module.css';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';

export default function DuplicatesSection() {
	const { analyticsData } = useAnalytics();
	const { duplicates } = analyticsData;

	return (
		<div className={styles.section}>
			<div className={styles.cardsGrid}>
				<div className={styles.card}>
					<div className={styles.cardContent}>
						<div className={styles.iconContainer}>
							<Copy className={styles.icon} />
						</div>
						<div className={styles.info}>
							<div className={styles.label}>Arquivos Duplicados</div>
							<div className={styles.value}>{duplicates.totalCount.toLocaleString()}</div>
						</div>
					</div>
				</div>

				<div className={styles.card}>
					<div className={styles.cardContent}>
						<div className={styles.iconContainer}>
							<HardDrive className={styles.icon} />
						</div>
						<div className={styles.info}>
							<div className={styles.label}>Espaço Desperdiçado</div>
							<div className={styles.value}>{duplicates.wastedSpace}</div>
						</div>
					</div>
				</div>
			</div>

			<Card title='Maiores Duplicatas'>
				<div className={styles.tableContainer}>
					<table className={styles.table}>
						<thead>
							<tr>
								<th>Nome</th>
								<th>Tamanho</th>
								<th>Cópias</th>
								<th>Caminhos</th>
							</tr>
						</thead>
						<tbody>
							{duplicates.items.map((duplicate, index) => (
								<tr key={index}>
									<td className={styles.nameCell}>{duplicate.name}</td>
									<td className={styles.sizeCell}>{duplicate.size}</td>
									<td className={styles.copiesCell}>{duplicate.copies}</td>
									<td className={styles.pathsCell}>
										{duplicate.paths.slice(0, 2).map((path, i) => (
											<div key={i} className={styles.pathItem}>
												{path}
											</div>
										))}
										{duplicate.paths.length > 2 && (
											<div className={styles.moreItems}>+{duplicate.paths.length - 2} mais</div>
										)}
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
