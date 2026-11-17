import { useMemo, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import { invalidateMusicMetadataViews } from '@/features/music/queryKeys/invalidateMusicMetadataViews';
import { readRefusedMetadataField, updateTrackMetadata } from '@/service/music';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import type { IMusicData, TrackMetadataFieldName } from '@/types/music';
import {
    buildChangedFields,
    readInitialFormValues,
    type TextualFieldName,
    type TrackMetadataFormValues,
} from './trackMetadataFormValues';

type FieldErrors = Partial<Record<TrackMetadataFieldName, string>>;

interface UseTrackMetadataFormOptions {
    track: IMusicData;
    onSaved: () => void;
}

export default function useTrackMetadataForm({ track, onSaved }: UseTrackMetadataFormOptions) {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const queryClient = useQueryClient();
    const initialValues = useMemo(() => readInitialFormValues(track), [track]);
    const [values, setValues] = useState<TrackMetadataFormValues>(initialValues);
    const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
    const [topError, setTopError] = useState<string | undefined>(undefined);

    const { fields: changedFields, invalidNumericFields } = useMemo(
        () => buildChangedFields(initialValues, values),
        [initialValues, values]
    );

    const saveMutation = useMutation({
        mutationFn: () => updateTrackMetadata(track.id, changedFields),
        onMutate: () => {
            setFieldErrors({});
            setTopError(undefined);
        },
        onSuccess: () => {
            void invalidateMusicMetadataViews(queryClient);
            enqueueSnackbar(t('MUSIC_METADATA_SAVED'), { variant: 'success' });
            onSaved();
        },
        onError: (error) => {
            const message = extractBackendErrorMessage(error) ?? t('MUSIC_METADATA_SAVE_FAILED');
            const refusedField = readRefusedMetadataField(error);
            if (refusedField) {
                setFieldErrors({ [refusedField]: message });
                return;
            }
            setTopError(message);
        },
    });

    const clearFieldError = (fieldName: TrackMetadataFieldName) =>
        setFieldErrors((currentErrors) => ({ ...currentErrors, [fieldName]: undefined }));

    const setTextualValue = (fieldName: TextualFieldName, fieldValue: string) => {
        setValues((currentValues) => ({ ...currentValues, [fieldName]: fieldValue }));
        clearFieldError(fieldName);
    };

    const setGenres = (genres: string[]) => {
        setValues((currentValues) => ({ ...currentValues, genres }));
        clearFieldError('genres');
    };

    const submit = () => {
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
        saveMutation.mutate();
    };

    const hasChanges = Object.keys(changedFields).length > 0 || invalidNumericFields.length > 0;

    return {
        values,
        fieldErrors,
        topError,
        isSaving: saveMutation.isPending,
        canSave: hasChanges && !saveMutation.isPending,
        setTextualValue,
        setGenres,
        submit,
    };
}
