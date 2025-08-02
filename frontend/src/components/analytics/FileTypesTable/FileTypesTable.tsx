import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import Card from '../../ui/Card/Card';
import styles from './FileTypesTable.module.css';
import { formatSize } from '@/utils';

export default function FileTypesTable() {
	const { analyticsData } = useAnalytics();
	const { fileTypes } = analyticsData;

	return (
		<Card title='Resumo por Tipo de Arquivo'>
			<div className={styles.tableContainer}>
				<table className={styles.table}>
					<thead>
						<tr>
							<th>Tipo</th>
							<th>Quantidade</th>
							<th>Espaço Total</th>
							<th>Percentual</th>
						</tr>
					</thead>
					<tbody>
						{fileTypes.map((type) => (
							<tr key={type.format}>
								<td className={styles.typeCell}>{type.format}</td>
								<td>{type.total.toLocaleString()}</td>
								<td>{formatSize(type.size)}</td>
								<td>
									<div className={styles.percentageContainer}>
										<div className={styles.percentageBar}>
											<div className={styles.percentageFill} style={{ width: `${type.percentage}%` }}></div>
										</div>
										<span className={styles.percentageText}>{type.percentage.toPrecision(2)}%</span>
									</div>
								</td>
							</tr>
						))}
					</tbody>
				</table>
			</div>
		</Card>
	);
}
