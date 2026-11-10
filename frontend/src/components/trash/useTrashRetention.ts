import useI18n from '@/components/i18n/provider/i18nContext';
import { getTrashRetention, updateTrashRetention } from '@/service/trash';
import { trashQueryKey } from '@/shared/queryKeys/fileQueryKeys';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import type { TrashRetention } from '@/types/trash';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import { useCallback, useState } from 'react';

const retentionQueryKey = [trashQueryKey, 'retention'];

const formatRetentionDraft = (savedDays: number | undefined): string =>
    savedDays === undefined ? '' : String(savedDays);

const isPositiveWholeNumber = (draftDays: string): boolean => /^[1-9]\d*$/.test(draftDays.trim());

export const useTrashRetention = () => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const queryClient = useQueryClient();

    const retentionQuery = useQuery<TrashRetention>({
        queryKey: retentionQueryKey,
        queryFn: getTrashRetention,
        retry: false,
    });
    const savedDays = retentionQuery.data?.days;

    const [draftDays, setDraftDays] = useState(formatRetentionDraft(savedDays));
    const [draftSourceDays, setDraftSourceDays] = useState(savedDays);
    if (draftSourceDays !== savedDays) {
        setDraftSourceDays(savedDays);
        setDraftDays(formatRetentionDraft(savedDays));
    }

    const retentionMutation = useMutation({
        mutationFn: updateTrashRetention,
        onSuccess: (_savedRetention, requestedDays) => {
            queryClient.setQueryData<TrashRetention>(retentionQueryKey, { days: requestedDays });
            void queryClient.invalidateQueries({ queryKey: [trashQueryKey] });
        },
    });

    const isDraftBlank = draftDays.trim() === '';
    const isDraftInvalid = !isDraftBlank && !isPositiveWholeNumber(draftDays);

    const commitDraft = useCallback(async () => {
        if (isDraftBlank) {
            setDraftDays(formatRetentionDraft(savedDays));
            return;
        }
        if (isDraftInvalid) {
            return;
        }
        const requestedDays = Number(draftDays.trim());
        if (requestedDays === savedDays) {
            return;
        }
        try {
            await retentionMutation.mutateAsync(requestedDays);
            enqueueSnackbar(t('TRASH_RETENTION_SAVED'), { variant: 'success' });
        } catch (error) {
            enqueueSnackbar(extractBackendErrorMessage(error) ?? t('TRASH_ACTION_ERROR'), {
                variant: 'error',
            });
        }
    }, [draftDays, enqueueSnackbar, isDraftBlank, isDraftInvalid, retentionMutation, savedDays, t]);

    return {
        t,
        draftDays,
        changeDraft: setDraftDays,
        commitDraft,
        isDraftInvalid,
        isRetentionLoading: retentionQuery.isLoading,
        isRetentionUnknown: retentionQuery.isError && savedDays === undefined,
        isSaving: retentionMutation.isPending,
    };
};
