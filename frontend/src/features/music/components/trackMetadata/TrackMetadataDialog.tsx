import {
    Alert,
    Button,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    Stack,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { IMusicData } from '@/types/music';
import TrackMetadataFieldsEditor from './TrackMetadataFieldsEditor';
import useTrackMetadataForm from './useTrackMetadataForm';
import useTrackMetadataSuggestions from './useTrackMetadataSuggestions';

interface TrackMetadataDialogProps {
    track: IMusicData;
    onClose: () => void;
}

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
                    <TrackMetadataFieldsEditor
                        values={form.values}
                        fieldErrors={form.fieldErrors}
                        isDisabled={form.isSaving}
                        artistNames={artistNames}
                        genreNames={genreNames}
                        onTextualChange={form.setTextualValue}
                        onGenresChange={form.setGenres}
                    />
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
