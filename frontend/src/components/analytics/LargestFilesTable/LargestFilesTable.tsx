import { File } from 'lucide-react';

import Card from '../../ui/Card/Card';
import styles from './LargestFilesTable.module.css';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import { formatSize } from '@/utils';

export default function LargestFilesTable() {
	const { analyticsData } = useAnalytics();
	const { largestFiles } = analyticsData;

	return (
		<Card title='Maiores Arquivos'>
			<div className={styles.tableContainer}>
				<table className={styles.table}>
					<thead>
						<tr>
							<th>Arquivo</th>
							<th>Tamanho</th>
							<th>Caminho</th>
						</tr>
					</thead>
					<tbody>
						{largestFiles.map((file, index) => (
							<tr key={index}>
								<td>
									<div className={styles.fileCell}>
										<File className={styles.fileIcon} />
										<span className={styles.fileName}>{file.name}</span>
									</div>
								</td>
								<td className={styles.sizeCell}>{formatSize(file.size)}</td>
								<td className={styles.pathCell}>{file.path}</td>
							</tr>
						))}
					</tbody>
				</table>
			</div>
		</Card>
	);
}
