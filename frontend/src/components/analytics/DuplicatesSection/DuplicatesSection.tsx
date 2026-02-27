import { Box, Card, CardContent, CardHeader, Grid, Table, TableBody, TableCell, TableHead, TableRow, Typography } from '@mui/material';
import { Copy, HardDrive } from 'lucide-react';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import { formatSize } from '@/utils';

export default function DuplicatesSection() {
	const { analyticsData } = useAnalytics();
	const { duplicates } = analyticsData;

	if (!duplicates || duplicates?.files?.length === 0) {
		return (
			<Card>
				<CardHeader title='Arquivos Duplicados' titleTypographyProps={{ variant: 'h6' }} />
				<CardContent>
					<Typography>Nenhum arquivo duplicado encontrado.</Typography>
				</CardContent>
			</Card>
		);
	}

	return (
		<Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
			<Grid container spacing={2}>
				<Grid size={{ xs: 12, sm: 6 }}>
					<Card>
						<CardContent sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
							<Copy size={32} />
							<Box>
								<Typography variant='body2' color='text.secondary'>Arquivos Duplicados</Typography>
								<Typography variant='h6'>{duplicates.total}</Typography>
							</Box>
						</CardContent>
					</Card>
				</Grid>
				<Grid size={{ xs: 12, sm: 6 }}>
					<Card>
						<CardContent sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
							<HardDrive size={32} />
							<Box>
								<Typography variant='body2' color='text.secondary'>Espaço Desperdiçado</Typography>
								<Typography variant='h6'>{formatSize(duplicates.total_size)}</Typography>
							</Box>
						</CardContent>
					</Card>
				</Grid>
			</Grid>

			<Card>
				<CardHeader title='Maiores Duplicatas' titleTypographyProps={{ variant: 'h6' }} />
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
								<TableCell component='th' scope='row'>{row.name}</TableCell>
								<TableCell align='right'>{formatSize(row.size)}</TableCell>
								<TableCell align='right'>{row.copies}</TableCell>
								<TableCell align='right'>
									{row.paths.slice(0, 2).map((path, i) => (
										<Box key={i} sx={{ fontSize: '0.75rem' }}>{path}</Box>
									))}
									{row.paths.length > 2 && (
										<Box sx={{ fontSize: '0.75rem', color: 'text.secondary' }}>+{row.paths.length - 2} mais</Box>
									)}
								</TableCell>
							</TableRow>
						))}
					</TableBody>
				</Table>
			</Card>
		</Box>
	);
}
