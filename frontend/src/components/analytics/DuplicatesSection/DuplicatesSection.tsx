import { Box, Card, CardContent, CardHeader, Grid, Table, TableBody, TableCell, TableHead, TableRow, Typography } from '@mui/material';
import { Copy, HardDrive } from 'lucide-react';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import useI18n from '@/components/i18n/provider/i18nContext';
import { formatSize } from '@/utils';

export default function DuplicatesSection() {
	const { analyticsData } = useAnalytics();
	const { t } = useI18n();
	const { duplicates } = analyticsData;

	if (!duplicates || duplicates?.files?.length === 0) {
		return (
			<Card>
				<CardHeader title={t('ANALYTICS_DUPLICATE_FILES')} titleTypographyProps={{ variant: 'h6' }} />
				<CardContent>
					<Typography>{t('ANALYTICS_NO_DUPLICATES')}</Typography>
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
								<Typography variant='body2' color='text.secondary'>{t('ANALYTICS_DUPLICATE_FILES')}</Typography>
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
								<Typography variant='body2' color='text.secondary'>{t('ANALYTICS_WASTED_SPACE')}</Typography>
								<Typography variant='h6'>{formatSize(duplicates.total_size)}</Typography>
							</Box>
						</CardContent>
					</Card>
				</Grid>
			</Grid>

			<Card>
				<CardHeader title={t('ANALYTICS_LARGEST_DUPLICATES')} titleTypographyProps={{ variant: 'h6' }} />
				<Table sx={{ minWidth: 650 }} aria-label='simple table'>
					<TableHead>
						<TableRow>
							<TableCell>{t('NAME')}</TableCell>
							<TableCell align='right'>{t('ANALYTICS_FILE_SIZE')}</TableCell>
							<TableCell align='right'>{t('ANALYTICS_COPIES')}</TableCell>
							<TableCell align='right'>{t('ANALYTICS_PATHS')}</TableCell>
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
										<Box sx={{ fontSize: '0.75rem', color: 'text.secondary' }}>+{row.paths.length - 2} {t('ANALYTICS_MORE')}</Box>
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
