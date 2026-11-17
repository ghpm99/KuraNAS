import {
    Alert,
    Autocomplete,
    Button,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    Stack,
    TextField,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { IMusicData } from '@/types/music';
import SuggestionTextField from './SuggestionTextField';
import useTrackMetadataForm from './useTrackMetadataForm';
import useTrackMetadataSuggestions from './useTrackMetadataSuggestions';
import type { NumericFieldName } from './trackMetadataFormValues';

interface TrackMetadataDialogProps {
    track: IMusicData;
    onClose: () => void;
}

const numericFields: { name: NumericFieldName; labelKey: string }[] = [
    { name: 'year', labelKey: 'MUSIC_METADATA_FIELD_YEAR' },
    { name: 'disc_number', labelKey: 'MUSIC_METADATA_FIELD_DISC_NUMBER' },
    { name: 'track_number', labelKey: 'MUSIC_METADATA_FIELD_TRACK_NUMBER' },
];

export default function TrackMetadataDialog({ track, onClose }: TrackMetadataDialogProps) {
    const { t } = useI18n();
    const form = useTrackMetadataForm({ track, onSaved: onClose });
    const { artistNames, genreNames } = useTrackMetadataSuggestions();

    return (
        <Dialog open onClose={form.isSaving ? undefined : onClose} fullWidth maxWidth="sm">
            <DialogTitle>{t('MUSIC_METADATA_DIALOG_TITLE')}</DialogTitle>
            <DialogContent>
                <Stack spacing={2} sx={{ pt: 1 }}>
                    {form.topError && <Alert severity="error">{form.topError}</Alert>}
                    <TextField
                        label={t('MUSIC_METADATA_FIELD_TITLE')}
                        value={form.values.title}
                        onChange={(event) => form.setTextualValue('title', event.target.value)}
                        error={form.fieldErrors.title !== undefined}
                        helperText={form.fieldErrors.title}
                        disabled={form.isSaving}
                    />
                    <SuggestionTextField
                        label={t('MUSIC_METADATA_FIELD_ARTIST')}
                        value={form.values.artist}
                        suggestions={artistNames}
                        errorMessage={form.fieldErrors.artist}
                        isDisabled={form.isSaving}
                        onChange={(nextValue) => form.setTextualValue('artist', nextValue)}
                    />
                    <TextField
                        label={t('MUSIC_METADATA_FIELD_ALBUM')}
                        value={form.values.album}
                        onChange={(event) => form.setTextualValue('album', event.target.value)}
                        error={form.fieldErrors.album !== undefined}
                        helperText={form.fieldErrors.album}
                        disabled={form.isSaving}
                    />
                    <SuggestionTextField
                        label={t('MUSIC_METADATA_FIELD_ALBUM_ARTIST')}
                        value={form.values.album_artist}
                        suggestions={artistNames}
                        errorMessage={form.fieldErrors.album_artist}
                        isDisabled={form.isSaving}
                        onChange={(nextValue) => form.setTextualValue('album_artist', nextValue)}
                    />
                    <Autocomplete
                        multiple
                        freeSolo
                        autoSelect
                        options={genreNames}
                        value={form.values.genres}
                        onChange={(_event, nextGenres) => form.setGenres(nextGenres)}
                        disabled={form.isSaving}
                        renderInput={(inputProps) => (
                            <TextField
                                {...inputProps}
                                label={t('MUSIC_METADATA_FIELD_GENRES')}
                                error={form.fieldErrors.genres !== undefined}
                                helperText={form.fieldErrors.genres}
                            />
                        )}
                    />
                    {numericFields.map(({ name, labelKey }) => (
                        <TextField
                            key={name}
                            label={t(labelKey)}
                            value={form.values[name]}
                            onChange={(event) => form.setTextualValue(name, event.target.value)}
                            error={form.fieldErrors[name] !== undefined}
                            helperText={form.fieldErrors[name]}
                            disabled={form.isSaving}
                            slotProps={{ htmlInput: { inputMode: 'numeric' } }}
                        />
                    ))}
                </Stack>
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose} disabled={form.isSaving}>
                    {t('ACTION_CANCEL')}
                </Button>
                <Button
                    variant="contained"
                    onClick={form.submit}
                    disabled={!form.canSave}
                    startIcon={form.isSaving ? <CircularProgress size={16} /> : undefined}
                >
                    {t('MUSIC_METADATA_SAVE')}
                </Button>
            </DialogActions>
        </Dialog>
    );
}
