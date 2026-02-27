import { useAbout } from '@/components/hooks/AboutProvider/AboutContext';
import useI18n from '@/components/i18n/provider/i18nContext';
import Card from '@/components/ui/Card/Card';
import { Box, Chip, Divider, Stack, Typography } from '@mui/material';

const StatusCard = () => {
	const { enable_workers, path, uptime } = useAbout();
	const { t } = useI18n();

	return (
		<Card title={t('STATUS_SYSTEM_TITLE')}>
			<Stack divider={<Divider />} spacing={2}>
				<Box>
					<Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 0.5 }}>
						<Typography variant='body2' fontWeight={500}>{t('WORKERS')}</Typography>
						<Chip
							label={enable_workers ? t('ENABLED_WORKERS') : t('DISABLED_WORKERS')}
							color={enable_workers ? 'success' : 'default'}
							size='small'
						/>
					</Box>
					<Typography variant='caption' color='text.secondary'>
						{enable_workers ? t('ENABLED_WORKERS_DESCRIPTION') : t('DISABLED_WORKERS_DESCRIPTION')}
					</Typography>
				</Box>

				<Box>
					<Typography variant='body2' fontWeight={500} gutterBottom>{t('WATCH_PATH')}</Typography>
					<Typography variant='body2' sx={{ fontFamily: 'monospace', mb: 0.5 }}>{path}</Typography>
					<Typography variant='caption' color='text.secondary'>{t('WATCH_PATH_DESCRIPTION')}</Typography>
				</Box>

				<Box>
					<Typography variant='body2' fontWeight={500} gutterBottom>{t('UPTIME')}</Typography>
					<Typography variant='h6' gutterBottom>{uptime}</Typography>
					<Typography variant='caption' color='text.secondary'>{t('UPTIME_DESCRIPTION')}</Typography>
				</Box>
			</Stack>
		</Card>
	);
};

export default StatusCard;
