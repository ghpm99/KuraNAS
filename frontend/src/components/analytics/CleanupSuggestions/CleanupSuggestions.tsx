import { Alert, Box, Card, CardHeader, List, ListItem, ListItemIcon, ListItemText } from '@mui/material';
import { FileX, Trash2 } from 'lucide-react';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';

export default function CleanupSuggestions() {
	const { analyticsData } = useAnalytics();
	const { cleanup } = analyticsData;

	return (
		<Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
			{cleanup.criticalSpace && (
				<Alert severity='error'>
					<strong>Espaço Crítico</strong>
					<br />
					O uso de disco ultrapassou 90%. Considere limpar arquivos desnecessários.
				</Alert>
			)}

			<Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 2 }}>
				<Card>
					<CardHeader title='Arquivos Grandes Não Acessados' titleTypographyProps={{ variant: 'h6' }} />
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
					<CardHeader title='Arquivos com Nomes Similares' titleTypographyProps={{ variant: 'h6' }} />
					<List dense>
						{cleanup.similarNames.map((similar, index) => (
							<ListItem key={index}>
								<ListItemIcon sx={{ minWidth: 36 }}><Trash2 size={16} /></ListItemIcon>
								<ListItemText
									primary={`${similar.name1} / ${similar.name2}`}
									secondary={`Similaridade: ${similar.similarity}%`}
								/>
							</ListItem>
						))}
					</List>
				</Card>
			</Box>
		</Box>
	);
}
