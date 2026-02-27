import { Box, Card, CardContent, Grid, Typography } from '@mui/material';
import { Database, File, Folder, HardDrive } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';

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
	const { storageOverview } = analyticsData;

	return (
		<Grid container spacing={2}>
			<Grid size={{ xs: 12, sm: 6, md: 3 }}>
				<StatCard icon={HardDrive} label='Espaço Utilizado' value={storageOverview.totalUsedSpace} />
			</Grid>
			<Grid size={{ xs: 12, sm: 6, md: 3 }}>
				<StatCard icon={File} label='Arquivos Armazenados' value={storageOverview.totalFiles.toLocaleString()} />
			</Grid>
			<Grid size={{ xs: 12, sm: 6, md: 3 }}>
				<StatCard icon={Folder} label='Pastas Armazenadas' value={storageOverview.totalFolders.toLocaleString()} />
			</Grid>
			<Grid size={{ xs: 12, sm: 6, md: 3 }}>
				<StatCard icon={Database} label='Espaço Livre' value={storageOverview.availableSpace} />
			</Grid>
		</Grid>
	);
}
