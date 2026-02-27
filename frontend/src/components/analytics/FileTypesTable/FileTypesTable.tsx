import { Box, Card, CardHeader, LinearProgress, Table, TableBody, TableCell, TableHead, TableRow } from '@mui/material';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';
import { formatSize } from '@/utils';

export default function FileTypesTable() {
	const { analyticsData } = useAnalytics();
	const { fileTypes } = analyticsData;

	return (
		<Card>
			<CardHeader title='Resumo por Tipo de Arquivo' titleTypographyProps={{ variant: 'h6' }} />
			<Table>
				<TableHead>
					<TableRow>
						<TableCell>Tipo</TableCell>
						<TableCell>Quantidade</TableCell>
						<TableCell>Espaço Total</TableCell>
						<TableCell>Percentual</TableCell>
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
