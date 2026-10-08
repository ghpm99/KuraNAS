import useI18n from '@/components/i18n/provider/i18nContext';
import { extractBackendErrorMessage } from '@/features/files/fileActions/bulkOutcome';
import { promoteFileToHot } from '@/service/tiering';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';

const usePromoteFileToHot = (fileId: number | undefined) => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const queryClient = useQueryClient();

    const { mutate, isPending } = useMutation({
        mutationFn: () => promoteFileToHot(fileId as number),
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: ['files'] });
            enqueueSnackbar(t('FILE_PROMOTE_TO_HOT_SUCCESS'), { variant: 'success' });
        },
        onError: (error) => {
            enqueueSnackbar(extractBackendErrorMessage(error) ?? t('ERROR_PROMOTE_TO_HOT_FAILED'), {
                variant: 'error',
            });
        },
    });

    return { promoteToHot: () => mutate(), isPromoting: isPending };
};

export default usePromoteFileToHot;
