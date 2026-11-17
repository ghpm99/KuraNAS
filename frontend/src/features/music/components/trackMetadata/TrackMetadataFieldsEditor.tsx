import { Autocomplete, Stack, TextField } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { TrackMetadataFieldName } from '@/types/music';
import SuggestionTextField from './SuggestionTextField';
import type {
    NumericFieldName,
    TextualFieldName,
    TrackMetadataFormValues,
} from './trackMetadataFormValues';

interface TrackMetadataFieldsEditorProps {
    values: TrackMetadataFormValues;
    fieldErrors: Partial<Record<TrackMetadataFieldName, string>>;
    isDisabled: boolean;
    artistNames: string[];
    genreNames: string[];
    onTextualChange: (fieldName: TextualFieldName, fieldValue: string) => void;
    onGenresChange: (genres: string[]) => void;
}

const numericFields: { name: NumericFieldName; labelKey: string }[] = [
    { name: 'year', labelKey: 'MUSIC_METADATA_FIELD_YEAR' },
    { name: 'disc_number', labelKey: 'MUSIC_METADATA_FIELD_DISC_NUMBER' },
    { name: 'track_number', labelKey: 'MUSIC_METADATA_FIELD_TRACK_NUMBER' },
];

export default function TrackMetadataFieldsEditor({
    values,
    fieldErrors,
    isDisabled,
    artistNames,
    genreNames,
    onTextualChange,
    onGenresChange,
}: TrackMetadataFieldsEditorProps) {
    const { t } = useI18n();

    return (
        <Stack spacing={2}>
            <TextField
                label={t('MUSIC_METADATA_FIELD_TITLE')}
                value={values.title}
                onChange={(event) => onTextualChange('title', event.target.value)}
                error={fieldErrors.title !== undefined}
                helperText={fieldErrors.title}
                disabled={isDisabled}
            />
            <SuggestionTextField
                label={t('MUSIC_METADATA_FIELD_ARTIST')}
                value={values.artist}
                suggestions={artistNames}
                errorMessage={fieldErrors.artist}
                isDisabled={isDisabled}
                onChange={(nextValue) => onTextualChange('artist', nextValue)}
            />
            <TextField
                label={t('MUSIC_METADATA_FIELD_ALBUM')}
                value={values.album}
                onChange={(event) => onTextualChange('album', event.target.value)}
                error={fieldErrors.album !== undefined}
                helperText={fieldErrors.album}
                disabled={isDisabled}
            />
            <SuggestionTextField
                label={t('MUSIC_METADATA_FIELD_ALBUM_ARTIST')}
                value={values.album_artist}
                suggestions={artistNames}
                errorMessage={fieldErrors.album_artist}
                isDisabled={isDisabled}
                onChange={(nextValue) => onTextualChange('album_artist', nextValue)}
            />
            <Autocomplete
                multiple
                freeSolo
                autoSelect
                options={genreNames}
                value={values.genres}
                onChange={(_event, nextGenres) => onGenresChange(nextGenres)}
                disabled={isDisabled}
                renderInput={(inputProps) => (
                    <TextField
                        {...inputProps}
                        label={t('MUSIC_METADATA_FIELD_GENRES')}
                        error={fieldErrors.genres !== undefined}
                        helperText={fieldErrors.genres}
                    />
                )}
            />
            {numericFields.map(({ name, labelKey }) => (
                <TextField
                    key={name}
                    label={t(labelKey)}
                    value={values[name]}
                    onChange={(event) => onTextualChange(name, event.target.value)}
                    error={fieldErrors[name] !== undefined}
                    helperText={fieldErrors[name]}
                    disabled={isDisabled}
                    slotProps={{ htmlInput: { inputMode: 'numeric' } }}
                />
            ))}
        </Stack>
    );
}
