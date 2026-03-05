import { Box, Card, CardHeader, LinearProgress, Table, TableBody, TableCell, TableHead, TableRow } from '@mui/material';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import useI18n from '@/components/i18n/provider/i18nContext';
import { formatSize } from '@/utils';

export default function FileTypesTable() {
	const { analyticsData } = useAnalytics();
	const { t } = useI18n();
	const { fileTypes } = analyticsData;

	return (
		<Card>
			<CardHeader title={t('ANALYTICS_FILE_TYPE_SUMMARY')} titleTypographyProps={{ variant: 'h6' }} />
			<Table>
				<TableHead>
					<TableRow>
						<TableCell>{t('ANALYTICS_FILE_TYPE')}</TableCell>
						<TableCell>{t('ANALYTICS_FILE_COUNT')}</TableCell>
						<TableCell>{t('ANALYTICS_TOTAL_SPACE')}</TableCell>
						<TableCell>{t('ANALYTICS_PERCENTAGE')}</TableCell>
					</TableRow>
				</TableHead>
				<TableBody>
					{fileTypes.map((type) => (
						<TableRow key={type.format}>
							<TableCell>{type.format}</TableCell>
							<TableCell>{type.total.toLocaleString()}</TableCell>
							<TableCell>{formatSize(type.size)}</TableCell>
							<TableCell sx={{ minWidth: 140 }}>
								<LinearProgress variant='determinate' value={type.percentage} sx={{ mb: 0.5 }} />
								<Box component='span' sx={{ fontSize: '0.75rem' }}>{type.percentage.toPrecision(2)}%</Box>
							</TableCell>
						</TableRow>
					))}
				</TableBody>
			</Table>
		</Card>
	);
}
