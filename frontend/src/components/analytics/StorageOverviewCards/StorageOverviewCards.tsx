import { Box, Card, CardContent, Grid, Typography } from '@mui/material';
import { Database, File, Folder, HardDrive } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import useI18n from '@/components/i18n/provider/i18nContext';

function StatCard({ icon: Icon, label, value }: { icon: LucideIcon; label: string; value: string }) {
	return (
		<Card>
			<CardContent sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
				<Icon size={32} />
				<Box>
					<Typography variant='body2' color='text.secondary'>{label}</Typography>
					<Typography variant='h6'>{value}</Typography>
				</Box>
			</CardContent>
		</Card>
	);
}

export default function StorageOverviewCards() {
	const { analyticsData } = useAnalytics();
	const { t } = useI18n();
	const { storageOverview } = analyticsData;

	return (
		<Grid container spacing={2}>
			<Grid size={{ xs: 12, sm: 6, md: 3 }}>
				<StatCard icon={HardDrive} label={t('ANALYTICS_SPACE_USED')} value={storageOverview.totalUsedSpace} />
			</Grid>
			<Grid size={{ xs: 12, sm: 6, md: 3 }}>
				<StatCard icon={File} label={t('ANALYTICS_FILES_STORED')} value={storageOverview.totalFiles.toLocaleString()} />
			</Grid>
			<Grid size={{ xs: 12, sm: 6, md: 3 }}>
				<StatCard icon={Folder} label={t('ANALYTICS_FOLDERS_STORED')} value={storageOverview.totalFolders.toLocaleString()} />
			</Grid>
			<Grid size={{ xs: 12, sm: 6, md: 3 }}>
				<StatCard icon={Database} label={t('ANALYTICS_FREE_SPACE')} value={storageOverview.availableSpace} />
			</Grid>
		</Grid>
	);
}
