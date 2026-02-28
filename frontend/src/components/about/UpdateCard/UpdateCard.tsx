import { useAbout } from '@/components/hooks/AboutProvider/AboutContext';
import useI18n from '@/components/i18n/provider/i18nContext';
import Card from '@/components/ui/Card/Card';
import { applyUpdate, getUpdateStatus } from '@/service/update';
import { UpdateStatus } from '@/types/update';
import { Box, Button, Chip, CircularProgress, Divider, Stack, Typography } from '@mui/material';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';

const UpdateCard = () => {
	const { version } = useAbout();
	const { t } = useI18n();
	const { enqueueSnackbar } = useSnackbar();
	const queryClient = useQueryClient();

	const {
		data: updateStatus,
		isLoading,
		isError,
		error,
		refetch,
	} = useQuery<UpdateStatus>({
		queryKey: ['update-status'],
		queryFn: getUpdateStatus,
		enabled: false,
	});

	const updateMutation = useMutation({
		mutationFn: applyUpdate,
		onSuccess: () => {
			enqueueSnackbar(t('UPDATE_APPLIED_SUCCESS'), { variant: 'success' });
		},
		onError: () => {
			enqueueSnackbar(t('UPDATE_APPLIED_ERROR'), { variant: 'error' });
			queryClient.invalidateQueries({ queryKey: ['update-status'] });
		},
	});

	const formatFileSize = (bytes: number) => {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
	};

	const formatDate = (dateStr: string) => {
		if (!dateStr) return '';
		return new Date(dateStr).toLocaleDateString();
	};

	return (
		<Card title={t('UPDATE_CARD_TITLE')}>
			<Stack divider={<Divider />} spacing={2}>
				<Box>
					<Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 0.5 }}>
						<Typography variant='body2' fontWeight={500}>
							{t('CURRENT_VERSION')}
						</Typography>
						<Chip label={version} size='small' />
					</Box>
				</Box>

				{updateStatus && (
					<>
						<Box>
							<Box
								sx={{
									display: 'flex',
									justifyContent: 'space-between',
									alignItems: 'center',
									mb: 0.5,
								}}
							>
								<Typography variant='body2' fontWeight={500}>
									{t('LATEST_VERSION')}
								</Typography>
								<Chip
									label={updateStatus.latest_version}
									color={updateStatus.update_available ? 'warning' : 'success'}
									size='small'
								/>
							</Box>
							{updateStatus.release_date && (
								<Typography variant='caption' color='text.secondary'>
									{t('RELEASE_DATE')}: {formatDate(updateStatus.release_date)}
								</Typography>
							)}
						</Box>

						{updateStatus.update_available && updateStatus.asset_size > 0 && (
							<Box>
								<Typography variant='caption' color='text.secondary'>
									{t('DOWNLOAD_SIZE')}: {formatFileSize(updateStatus.asset_size)}
								</Typography>
							</Box>
						)}

						{updateStatus.release_notes && (
							<Box>
								<Typography variant='body2' fontWeight={500} gutterBottom>
									{t('RELEASE_NOTES')}
								</Typography>
								<Typography
									variant='body2'
									color='text.secondary'
									sx={{ whiteSpace: 'pre-wrap', maxHeight: 200, overflow: 'auto' }}
								>
									{updateStatus.release_notes}
								</Typography>
							</Box>
						)}
					</>
				)}

				{isError && (
					<Box>
						<Typography variant='body2' color='error'>
							{t('UPDATE_CHECK_ERROR')}: {error instanceof Error ? error.message : String(error)}
						</Typography>
					</Box>
				)}

				<Box sx={{ display: 'flex', gap: 1 }}>
					<Button
						variant='outlined'
						size='small'
						onClick={() => refetch()}
						disabled={isLoading || updateMutation.isPending}
						startIcon={isLoading ? <CircularProgress size={16} /> : undefined}
					>
						{t('CHECK_FOR_UPDATES')}
					</Button>

					{updateStatus?.update_available && (
						<Button
							variant='contained'
							size='small'
							color='primary'
							onClick={() => updateMutation.mutate()}
							disabled={updateMutation.isPending}
							startIcon={updateMutation.isPending ? <CircularProgress size={16} /> : undefined}
						>
							{t('UPDATE_NOW')}
						</Button>
					)}
				</Box>
			</Stack>
		</Card>
	);
};

export default UpdateCard;
