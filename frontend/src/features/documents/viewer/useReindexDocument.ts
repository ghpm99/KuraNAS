import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import { reindexDocument } from '@/service/documents';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { documentIndexStateQueryKey, documentTextQueryKey } from '../documentsQueryKeys';

const useReindexDocument = (fileId: number) => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: () => reindexDocument(fileId),
        onSuccess: async () => {
            enqueueSnackbar(t('DOCUMENTS_REINDEX_SUCCESS'), { variant: 'success' });
            await queryClient.invalidateQueries({ queryKey: documentIndexStateQueryKey(fileId) });
            await queryClient.invalidateQueries({ queryKey: documentTextQueryKey(fileId) });
        },
        onError: (error) => {
            enqueueSnackbar(extractBackendErrorMessage(error) || t('DOCUMENTS_REINDEX_ERROR'), {
                variant: 'error',
            });
        },
    });
};

export default useReindexDocument;
