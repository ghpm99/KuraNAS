import { Card, CardContent, Typography } from '@mui/material';

interface AnalyticsKpiCardProps {
	title: string;
	value: string;
	helpText?: string;
}

export default function AnalyticsKpiCard({ title, value, helpText }: AnalyticsKpiCardProps) {
	return (
		<Card>
			<CardContent>
				<Typography variant='body2' color='text.secondary'>
					{title}
				</Typography>
				<Typography variant='h5' sx={{ mt: 1 }}>
					{value}
				</Typography>
				{helpText ? (
					<Typography variant='caption' color='text.secondary'>
						{helpText}
					</Typography>
				) : null}
			</CardContent>
		</Card>
	);
}
