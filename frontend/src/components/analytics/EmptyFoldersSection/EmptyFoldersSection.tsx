import { Box, Card, CardContent, CardHeader, List, ListItem, ListItemIcon, ListItemText, Typography } from '@mui/material';
import { FolderX } from 'lucide-react';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';

export default function EmptyFoldersSection() {
	const { analyticsData } = useAnalytics();
	const { organization } = analyticsData;

	return (
		<Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
			<Card>
				<CardContent sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
					<FolderX size={32} />
					<Box>
						<Typography variant='body2' color='text.secondary'>Pastas Vazias</Typography>
						<Typography variant='h6'>{organization.emptyFolders}</Typography>
					</Box>
				</CardContent>
			</Card>

			<Card>
				<CardHeader title='Caminhos Vazios' titleTypographyProps={{ variant: 'h6' }} />
				<List dense>
					{organization.emptyPaths.map((path, index) => (
						<ListItem key={index}>
							<ListItemIcon sx={{ minWidth: 36 }}><FolderX size={16} /></ListItemIcon>
							<ListItemText primary={path} />
						</ListItem>
					))}
				</List>
			</Card>
		</Box>
	);
}
