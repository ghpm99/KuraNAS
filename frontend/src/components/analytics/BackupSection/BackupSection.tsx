import { Box, Card, CardContent, CardHeader, Chip, Grid, Table, TableBody, TableCell, TableHead, TableRow, Typography } from '@mui/material';
import { Calendar, HardDrive } from 'lucide-react';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import useI18n from '@/components/i18n/provider/i18nContext';

const statusColor: Record<string, 'success' | 'error' | 'warning'> = {
	success: 'success',
	failed: 'error',
	pending: 'warning',
};

export default function BackupSection() {
	const { analyticsData } = useAnalytics();
	const { t } = useI18n();
	const { backup } = analyticsData;

	const statusLabel: Record<string, string> = {
		success: t('ANALYTICS_BACKUP_SUCCESS'),
		failed: t('ANALYTICS_BACKUP_FAILED'),
		pending: t('ANALYTICS_BACKUP_PENDING'),
	};

	return (
		<Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
			<Grid container spacing={2}>
				<Grid size={{ xs: 12, sm: 6 }}>
					<Card>
						<CardContent sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
							<Calendar size={32} />
							<Box>
								<Typography variant='body2' color='text.secondary'>{t('ANALYTICS_LAST_BACKUP')}</Typography>
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
								<Typography variant='body2' color='text.secondary'>{t('ANALYTICS_LAST_BACKUP_SIZE')}</Typography>
								<Typography variant='h6'>{backup.lastBackupSize}</Typography>
							</Box>
						</CardContent>
					</Card>
				</Grid>
			</Grid>

			<Card>
				<CardHeader title={t('ANALYTICS_BACKUP_HISTORY')} titleTypographyProps={{ variant: 'h6' }} />
				<Table>
					<TableHead>
						<TableRow>
							<TableCell>{t('ANALYTICS_BACKUP_DATE')}</TableCell>
							<TableCell>{t('ANALYTICS_BACKUP_SIZE')}</TableCell>
							<TableCell>{t('ANALYTICS_BACKUP_STATUS')}</TableCell>
						</TableRow>
					</TableHead>
					<TableBody>
						{backup.history.map((item, index) => (
							<TableRow key={index}>
								<TableCell>{item.date}</TableCell>
								<TableCell>{item.size}</TableCell>
								<TableCell>
									<Chip
										label={statusLabel[item.status] ?? t('ANALYTICS_BACKUP_PENDING')}
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
