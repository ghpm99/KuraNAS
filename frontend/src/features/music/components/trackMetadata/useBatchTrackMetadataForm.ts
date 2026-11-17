import { useMemo, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import { invalidateMusicMetadataViews } from '@/features/music/queryKeys/invalidateMusicMetadataViews';
import { readRefusedMetadataField, updateTracksMetadata } from '@/service/music';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import type { BatchTrackMetadataResult, BatchTrackMetadataSelector } from '@/types/music';
import { buildFilledFields, emptyFormValues } from './trackMetadataFormValues';
import useTrackMetadataFieldState from './useTrackMetadataFieldState';

export type BatchTrackMetadataStep =
    | { kind: 'editing' }
    | { kind: 'reviewing'; matchedCount: number }
    | { kind: 'done'; result: BatchTrackMetadataResult };

interface UseBatchTrackMetadataFormOptions {
    selector: BatchTrackMetadataSelector;
    onClose: () => void;
}

const editingStep: BatchTrackMetadataStep = { kind: 'editing' };

export default function useBatchTrackMetadataForm({
    selector,
    onClose,
}: UseBatchTrackMetadataFormOptions) {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const queryClient = useQueryClient();
    const { values, fieldErrors, setFieldErrors, setTextualValue, setGenres } =
        useTrackMetadataFieldState(emptyFormValues);
    const [topError, setTopError] = useState<string | undefined>(undefined);
    const [step, setStep] = useState<BatchTrackMetadataStep>(editingStep);

    const { fields: filledFields, invalidNumericFields } = useMemo(
        () => buildFilledFields(values),
        [values]
    );

    const submitMutation = useMutation({
        mutationFn: (options: { isDryRun: boolean }) =>
            updateTracksMetadata(selector, filledFields, options),
        onMutate: () => {
            setFieldErrors({});
            setTopError(undefined);
        },
        onSuccess: (result, { isDryRun }) => {
            if (isDryRun) {
                setStep({ kind: 'reviewing', matchedCount: result.matched });
                return;
            }
            void invalidateMusicMetadataViews(queryClient);
            if (result.failed.length > 0) {
                setStep({ kind: 'done', result });
                return;
            }
            enqueueSnackbar(t('MUSIC_BATCH_METADATA_UPDATED', { count: String(result.updated) }), {
                variant: 'success',
            });
            onClose();
        },
        onError: (error) => {
            const message = extractBackendErrorMessage(error) ?? t('MUSIC_BATCH_METADATA_FAILED');
            const refusedField = readRefusedMetadataField(error);
            if (refusedField) {
                setFieldErrors({ [refusedField]: message });
                setStep(editingStep);
                return;
            }
            setTopError(message);
        },
    });

    const review = () => {
        if (invalidNumericFields.length > 0) {
            setFieldErrors(
                Object.fromEntries(
                    invalidNumericFields.map((fieldName) => [
                        fieldName,
                        t('MUSIC_METADATA_NUMBER_INVALID'),
                    ])
                )
            );
            return;
        }
        submitMutation.mutate({ isDryRun: true });
    };

    const confirm = () => submitMutation.mutate({ isDryRun: false });

    const backToEditing = () => {
        setTopError(undefined);
        setStep(editingStep);
    };

    const hasFilledFields = Object.keys(filledFields).length > 0 || invalidNumericFields.length > 0;

    return {
        values,
        fieldErrors,
        topError,
        step,
        isSubmitting: submitMutation.isPending,
        canReview: hasFilledFields && !submitMutation.isPending,
        setTextualValue,
        setGenres,
        review,
        confirm,
        backToEditing,
    };
}
