import { Box, Card, CardHeader, Table, TableBody, TableCell, TableHead, TableRow } from '@mui/material';
import { File } from 'lucide-react';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import useI18n from '@/components/i18n/provider/i18nContext';
import { formatSize } from '@/utils';

export default function LargestFilesTable() {
	const { analyticsData } = useAnalytics();
	const { t } = useI18n();
	const { largestFiles } = analyticsData;

	return (
		<Card>
			<CardHeader title={t('ANALYTICS_LARGEST_FILES')} titleTypographyProps={{ variant: 'h6' }} />
			<Table>
				<TableHead>
					<TableRow>
						<TableCell>{t('ANALYTICS_FILE')}</TableCell>
						<TableCell>{t('ANALYTICS_FILE_SIZE')}</TableCell>
						<TableCell>{t('ANALYTICS_FILE_PATH')}</TableCell>
					</TableRow>
				</TableHead>
				<TableBody>
					{largestFiles.map((file, index) => (
						<TableRow key={index}>
							<TableCell>
								<Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
									<File size={16} />
									{file.name}
								</Box>
							</TableCell>
							<TableCell>{formatSize(file.size)}</TableCell>
							<TableCell sx={{ color: 'text.secondary', fontSize: '0.75rem' }}>{file.path}</TableCell>
						</TableRow>
					))}
				</TableBody>
			</Table>
		</Card>
	);
}
