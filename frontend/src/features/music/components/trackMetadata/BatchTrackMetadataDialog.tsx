import {
    Alert,
    Button,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    List,
    ListItem,
    ListItemText,
    Stack,
    Typography,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { BatchTrackMetadataResult, BatchTrackMetadataSelector } from '@/types/music';
import TrackMetadataFieldsEditor from './TrackMetadataFieldsEditor';
import useBatchTrackMetadataForm from './useBatchTrackMetadataForm';
import useTrackMetadataSuggestions from './useTrackMetadataSuggestions';

interface BatchTrackMetadataDialogProps {
    selector: BatchTrackMetadataSelector;
    onClose: () => void;
}

function BatchFailureList({ failures }: { failures: BatchTrackMetadataResult['failed'] }) {
    const { t } = useI18n();

    return (
        <>
            <Typography variant="subtitle2">{t('MUSIC_BATCH_METADATA_FAILED_TITLE')}</Typography>
            <List dense disablePadding>
                {failures.map((failure) => (
                    <ListItem key={failure.fileId} disableGutters>
                        <ListItemText
                            primary={t('MUSIC_BATCH_METADATA_FAILED_FILE', {
                                id: String(failure.fileId),
                            })}
                            secondary={failure.error}
                        />
                    </ListItem>
                ))}
            </List>
        </>
    );
}

export default function BatchTrackMetadataDialog({
    selector,
    onClose,
}: BatchTrackMetadataDialogProps) {
    const { t } = useI18n();
    const form = useBatchTrackMetadataForm({ selector, onClose });
    const { artistNames, genreNames } = useTrackMetadataSuggestions();
    const titleKey = 'albumKey' in selector ? 'MUSIC_EDIT_ALBUM' : 'MUSIC_EDIT_FOLDER';
    const { step } = form;

    return (
        <Dialog open onClose={form.isSubmitting ? undefined : onClose} fullWidth maxWidth="sm">
            <DialogTitle>{t(titleKey)}</DialogTitle>
            <DialogContent>
                <Stack spacing={2} sx={{ pt: 1 }}>
                    {form.topError && <Alert severity="error">{form.topError}</Alert>}
                    {step.kind === 'editing' && (
                        <>
                            <Typography variant="body2" color="text.secondary">
                                {t('MUSIC_BATCH_METADATA_HINT')}
                            </Typography>
                            <TrackMetadataFieldsEditor
                                values={form.values}
                                fieldErrors={form.fieldErrors}
                                isDisabled={form.isSubmitting}
                                artistNames={artistNames}
                                genreNames={genreNames}
                                onTextualChange={form.setTextualValue}
                                onGenresChange={form.setGenres}
                            />
                        </>
                    )}
                    {step.kind === 'reviewing' && (
                        <Typography>
                            {step.matchedCount === 0
                                ? t('MUSIC_BATCH_METADATA_NO_MATCH')
                                : t('MUSIC_BATCH_METADATA_WILL_CHANGE', {
                                      count: String(step.matchedCount),
                                  })}
                        </Typography>
                    )}
                    {step.kind === 'done' && (
                        <>
                            <Typography>
                                {t('MUSIC_BATCH_METADATA_UPDATED', {
                                    count: String(step.result.updated),
                                })}
                            </Typography>
                            <BatchFailureList failures={step.result.failed} />
                        </>
                    )}
                </Stack>
            </DialogContent>
            <DialogActions>
                {step.kind === 'editing' && (
                    <>
                        <Button onClick={onClose} disabled={form.isSubmitting}>
                            {t('ACTION_CANCEL')}
                        </Button>
                        <Button
                            variant="contained"
                            onClick={form.review}
                            disabled={!form.canReview}
                            startIcon={
                                form.isSubmitting ? <CircularProgress size={16} /> : undefined
                            }
                        >
                            {t('MUSIC_BATCH_METADATA_REVIEW')}
                        </Button>
                    </>
                )}
                {step.kind === 'reviewing' && (
                    <>
                        <Button onClick={form.backToEditing} disabled={form.isSubmitting}>
                            {t('MUSIC_BATCH_METADATA_BACK')}
                        </Button>
                        <Button
                            variant="contained"
                            onClick={form.confirm}
                            disabled={step.matchedCount === 0 || form.isSubmitting}
                            startIcon={
                                form.isSubmitting ? <CircularProgress size={16} /> : undefined
                            }
                        >
                            {t('MUSIC_BATCH_METADATA_CONFIRM')}
                        </Button>
                    </>
                )}
                {step.kind === 'done' && <Button onClick={onClose}>{t('CLOSE')}</Button>}
            </DialogActions>
        </Dialog>
    );
}
