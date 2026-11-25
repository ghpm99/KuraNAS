import useI18n from '@/components/i18n/provider/i18nContext';
import { getYtDlpStatus, updateYtDlp } from '@/service/ingest';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';

const useYtDlpSettings = () => {
	const { t } = useI18n();
	const { enqueueSnackbar } = useSnackbar();
	const queryClient = useQueryClient();

	const statusQuery = useQuery({
		queryKey: ['ytdlp-status'],
		queryFn: getYtDlpStatus,
		retry: false,
	});

	const updateMutation = useMutation({
		mutationFn: updateYtDlp,
		onSuccess: () => {
			enqueueSnackbar(t('YTDLP_UPDATE_APPLIED'), { variant: 'success' });
			queryClient.invalidateQueries({ queryKey: ['ytdlp-status'] });
		},
		onError: (error: unknown) => {
			enqueueSnackbar(extractBackendErrorMessage(error) ?? t('ERROR_YTDLP_UPDATE'), { variant: 'error' });
		},
	});

	return {
		t,
		status: statusQuery.data,
		isLoading: statusQuery.isLoading,
		hasError: statusQuery.isError,
		loadErrorMessage: extractBackendErrorMessage(statusQuery.error) ?? t('CONNECTION_BANNER_UNAVAILABLE'),
		isUpdating: updateMutation.isPending,
		handleUpdate: () => updateMutation.mutate(),
	};
};

export default useYtDlpSettings;
