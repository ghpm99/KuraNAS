import { Box, Card, CardContent, CardHeader, Grid, Table, TableBody, TableCell, TableHead, TableRow, Typography } from '@mui/material';
import { File, Trash2 } from 'lucide-react';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';

export default function TrashSection() {
	const { analyticsData } = useAnalytics();
	const { trash } = analyticsData;

	return (
		<Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
			<Grid container spacing={2}>
				<Grid size={{ xs: 12, sm: 6 }}>
					<Card>
						<CardContent sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
							<File size={32} />
							<Box>
								<Typography variant='body2' color='text.secondary'>Arquivos na Lixeira</Typography>
								<Typography variant='h6'>{trash.totalFiles}</Typography>
							</Box>
						</CardContent>
					</Card>
				</Grid>
				<Grid size={{ xs: 12, sm: 6 }}>
					<Card>
						<CardContent sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
							<Trash2 size={32} />
							<Box>
								<Typography variant='body2' color='text.secondary'>Espaço Ocupado</Typography>
								<Typography variant='h6'>{trash.totalSpace}</Typography>
							</Box>
						</CardContent>
					</Card>
				</Grid>
			</Grid>

			<Card>
				<CardHeader title='Arquivos na Lixeira' titleTypographyProps={{ variant: 'h6' }} />
				<Table>
					<TableHead>
						<TableRow>
							<TableCell>Nome</TableCell>
							<TableCell>Tamanho</TableCell>
							<TableCell>Data de Exclusão</TableCell>
						</TableRow>
					</TableHead>
					<TableBody>
						{trash.files.map((file, index) => (
							<TableRow key={index}>
								<TableCell>
									<Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
										<Trash2 size={16} />
										{file.name}
									</Box>
								</TableCell>
								<TableCell>{file.size}</TableCell>
								<TableCell>{file.deletedDate}</TableCell>
							</TableRow>
						))}
					</TableBody>
				</Table>
			</Card>
		</Box>
	);
}
