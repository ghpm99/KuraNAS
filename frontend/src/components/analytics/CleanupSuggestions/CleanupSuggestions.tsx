import { Alert, Box, Card, CardHeader, List, ListItem, ListItemIcon, ListItemText } from '@mui/material';
import { FileX, Trash2 } from 'lucide-react';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import useI18n from '@/components/i18n/provider/i18nContext';

export default function CleanupSuggestions() {
	const { analyticsData } = useAnalytics();
	const { t } = useI18n();
	const { cleanup } = analyticsData;

	return (
		<Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
			{cleanup.criticalSpace && (
				<Alert severity='error'>
					<strong>{t('ANALYTICS_CRITICAL_SPACE')}</strong>
					<br />
					{t('ANALYTICS_CRITICAL_SPACE_MSG')}
				</Alert>
			)}

			<Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 2 }}>
				<Card>
					<CardHeader title={t('ANALYTICS_LARGE_UNACCESSED')} titleTypographyProps={{ variant: 'h6' }} />
					<List dense>
						{cleanup.oldLargeFiles.map((file, index) => (
							<ListItem key={index}>
								<ListItemIcon sx={{ minWidth: 36 }}><FileX size={16} /></ListItemIcon>
								<ListItemText primary={file.name} secondary={`${file.size} • ${file.path}`} />
							</ListItem>
						))}
					</List>
				</Card>

				<Card>
					<CardHeader title={t('ANALYTICS_SIMILAR_NAMES')} titleTypographyProps={{ variant: 'h6' }} />
					<List dense>
						{cleanup.similarNames.map((similar, index) => (
							<ListItem key={index}>
								<ListItemIcon sx={{ minWidth: 36 }}><Trash2 size={16} /></ListItemIcon>
								<ListItemText
									primary={`${similar.name1} / ${similar.name2}`}
									secondary={`${t('ANALYTICS_SIMILARITY')}${similar.similarity}%`}
								/>
							</ListItem>
						))}
					</List>
				</Card>
			</Box>
		</Box>
	);
}
