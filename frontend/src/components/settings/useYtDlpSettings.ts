import useI18n from '@/components/i18n/provider/i18nContext';
import { getYtDlpStatus, updateYtDlp } from '@/service/ingest';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';

const extractBackendError = (error: unknown): string | undefined => {
	if (typeof error === 'object' && error !== null && 'response' in error) {
		const response = (error as { response?: { data?: { error?: string } } }).response;
		return response?.data?.error;
	}
	return undefined;
};

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
			enqueueSnackbar(extractBackendError(error) ?? t('ERROR_YTDLP_UPDATE'), { variant: 'error' });
		},
	});

	return {
		t,
		status: statusQuery.data,
		isLoading: statusQuery.isLoading,
		hasError: statusQuery.isError,
		isUpdating: updateMutation.isPending,
		handleUpdate: () => updateMutation.mutate(),
	};
};

export default useYtDlpSettings;
