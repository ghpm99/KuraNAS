import { useState } from 'react';
import type { TrackMetadataFieldName } from '@/types/music';
import type { TextualFieldName, TrackMetadataFormValues } from './trackMetadataFormValues';

export type TrackMetadataFieldErrors = Partial<Record<TrackMetadataFieldName, string>>;

export default function useTrackMetadataFieldState(initialValues: TrackMetadataFormValues) {
    const [values, setValues] = useState<TrackMetadataFormValues>(initialValues);
    const [fieldErrors, setFieldErrors] = useState<TrackMetadataFieldErrors>({});

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

    return { values, fieldErrors, setFieldErrors, setTextualValue, setGenres };
}
