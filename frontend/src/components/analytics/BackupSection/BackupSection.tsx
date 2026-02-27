import { Box, Card, CardContent, CardHeader, Chip, Grid, Table, TableBody, TableCell, TableHead, TableRow, Typography } from '@mui/material';
import { Calendar, HardDrive } from 'lucide-react';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';

const statusColor: Record<string, 'success' | 'error' | 'warning'> = {
	success: 'success',
	failed: 'error',
	pending: 'warning',
};

const statusLabel: Record<string, string> = {
	success: 'Sucesso',
	failed: 'Falhou',
	pending: 'Pendente',
};

export default function BackupSection() {
	const { analyticsData } = useAnalytics();
	const { backup } = analyticsData;

	return (
		<Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
			<Grid container spacing={2}>
				<Grid size={{ xs: 12, sm: 6 }}>
					<Card>
						<CardContent sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
							<Calendar size={32} />
							<Box>
								<Typography variant='body2' color='text.secondary'>Último Backup</Typography>
								<Typography variant='h6'>{backup.lastBackup}</Typography>
							</Box>
						</CardContent>
					</Card>
				</Grid>
				<Grid size={{ xs: 12, sm: 6 }}>
					<Card>
						<CardContent sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
							<HardDrive size={32} />
							<Box>
								<Typography variant='body2' color='text.secondary'>Tamanho do Último Backup</Typography>
								<Typography variant='h6'>{backup.lastBackupSize}</Typography>
							</Box>
						</CardContent>
					</Card>
				</Grid>
			</Grid>

			<Card>
				<CardHeader title='Histórico de Backups' titleTypographyProps={{ variant: 'h6' }} />
				<Table>
					<TableHead>
						<TableRow>
							<TableCell>Data</TableCell>
							<TableCell>Tamanho</TableCell>
							<TableCell>Status</TableCell>
						</TableRow>
					</TableHead>
					<TableBody>
						{backup.history.map((item, index) => (
							<TableRow key={index}>
								<TableCell>{item.date}</TableCell>
								<TableCell>{item.size}</TableCell>
								<TableCell>
									<Chip
										label={statusLabel[item.status] ?? 'Pendente'}
										color={statusColor[item.status] ?? 'warning'}
										size='small'
									/>
								</TableCell>
							</TableRow>
						))}
					</TableBody>
				</Table>
			</Card>
		</Box>
	);
}
