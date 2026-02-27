import { Box, Card, CardHeader, List, ListItem, ListItemIcon, ListItemText } from '@mui/material';
import { Clock, Eye } from 'lucide-react';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';

export default function RecentActivity() {
	const { analyticsData } = useAnalytics();
	const { recentFiles, accessedFiles } = analyticsData.recentActivity;

	return (
		<Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
			<Card>
				<CardHeader title='Arquivos Recentemente Adicionados' titleTypographyProps={{ variant: 'h6' }} />
				<List dense>
					{recentFiles.map((file, index) => (
						<ListItem key={index} divider={index < recentFiles.length - 1}>
							<ListItemIcon sx={{ minWidth: 36 }}><Clock size={16} /></ListItemIcon>
							<ListItemText primary={file.name} secondary={`${file.size} • ${file.date}`} />
						</ListItem>
					))}
				</List>
			</Card>

			<Card>
				<CardHeader title='Arquivos Mais Acessados' titleTypographyProps={{ variant: 'h6' }} />
				<List dense>
					{accessedFiles.map((file, index) => (
						<ListItem key={index} divider={index < accessedFiles.length - 1}>
							<ListItemIcon sx={{ minWidth: 36 }}><Eye size={16} /></ListItemIcon>
							<ListItemText primary={file.name} secondary={`${file.accessCount} acessos • ${file.lastAccess}`} />
						</ListItem>
					))}
				</List>
			</Card>
		</Box>
	);
}
