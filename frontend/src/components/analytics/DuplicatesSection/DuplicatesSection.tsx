import { Copy, HardDrive } from 'lucide-react';

import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import { formatSize } from '@/utils';
import { Table, TableBody, TableCell, TableHead, TableRow } from '@mui/material';
import Card from '../../ui/Card/Card';
import styles from './DuplicatesSection.module.css';

export default function DuplicatesSection() {
	const { analyticsData } = useAnalytics();
	const { duplicates } = analyticsData;

	if (!duplicates || duplicates?.files?.length === 0) {
		return (
			<Card title='Arquivos Duplicados'>
				<div className={styles.emptyState}>
					<p>Nenhum arquivo duplicado encontrado.</p>
				</div>
			</Card>
		);
	}
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
							<div className={styles.value}>{duplicates.total}</div>
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
							<div className={styles.value}>{formatSize(duplicates.total_size)}</div>
						</div>
					</div>
				</div>
			</div>

			<Card title='Maiores Duplicatas'>
				<div className={styles.tableContainer}>
					<Table sx={{ minWidth: 650 }} aria-label='simple table'>
						<TableHead>
							<TableRow>
								<TableCell>Nome</TableCell>
								<TableCell align='right'>Tamanho</TableCell>
								<TableCell align='right'>Cópias</TableCell>
								<TableCell align='right'>Caminhos</TableCell>
							</TableRow>
						</TableHead>
						<TableBody>
							{duplicates?.files?.map((row) => (
								<TableRow key={row.name} sx={{ '&:last-child td, &:last-child th': { border: 0 } }}>
									<TableCell component='th' scope='row'>
										{row.name}
									</TableCell>
									<TableCell align='right'>{formatSize(row.size)}</TableCell>
									<TableCell align='right'>{row.copies}</TableCell>
									<TableCell align='right'>
										{row.paths.slice(0, 2).map((path, i) => (
											<div key={i} className={styles.pathItem}>
												{path}
											</div>
										))}
										{row.paths.length > 2 && <div className={styles.moreItems}>+{row.paths.length - 2} mais</div>}
									</TableCell>
								</TableRow>
							))}
						</TableBody>
					</Table>
				</div>
			</Card>
		</div>
	);
}
