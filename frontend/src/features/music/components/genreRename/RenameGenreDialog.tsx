import {
    Alert,
    Autocomplete,
    createFilterOptions,
    Button,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    Stack,
    TextField,
    Typography,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import useTrackMetadataSuggestions from '@/features/music/components/trackMetadata/useTrackMetadataSuggestions';
import type { MusicGenre } from '@/types/music';
import useRenameGenre, { type RenameGenreConfirmation } from './useRenameGenre';

interface RenameGenreDialogProps {
    genreKey: string;
    genreName: string;
    trackCount: number;
    onRenamed: (renamedGenreKey: string) => void;
    onClose: () => void;
}

const filterGenreOptionsByTypedName = createFilterOptions<MusicGenre>();

const describeGenreOption = (genreOption: string | MusicGenre) =>
    typeof genreOption === 'string' ? genreOption : genreOption.genre;

function RenameConfirmationMessage({
    genreName,
    confirmation,
}: {
    genreName: string;
    confirmation: RenameGenreConfirmation;
}) {
    const { t } = useI18n();
    const messageKey = confirmation.isMerge
        ? 'MUSIC_GENRE_RENAME_WILL_MERGE'
        : 'MUSIC_GENRE_RENAME_WILL_RENAME';

    return (
        <Typography>
            {t(messageKey, {
                count: String(confirmation.movedTrackCount),
                from: genreName,
                to: confirmation.targetName,
                total: String(confirmation.resultingTrackCount),
            })}
        </Typography>
    );
}

export default function RenameGenreDialog({
    genreKey,
    genreName,
    trackCount,
    onRenamed,
    onClose,
}: RenameGenreDialogProps) {
    const { t } = useI18n();
    const rename = useRenameGenre({ genreKey, genreName, trackCount, onRenamed, onClose });
    const { genreGroups } = useTrackMetadataSuggestions();
    const otherGenreGroups = genreGroups.filter((genreGroup) => genreGroup.key !== genreKey);

    return (
        <Dialog open onClose={rename.isSubmitting ? undefined : onClose} fullWidth maxWidth="xs">
            <DialogTitle>{t('MUSIC_GENRE_RENAME')}</DialogTitle>
            <DialogContent>
                <Stack spacing={2} sx={{ pt: 1 }}>
                    {rename.topError && <Alert severity="error">{rename.topError}</Alert>}
                    <Autocomplete
                        freeSolo
                        options={otherGenreGroups}
                        getOptionLabel={describeGenreOption}
                        filterOptions={(genreOptions, filterState) =>
                            rename.hasTypedName
                                ? filterGenreOptionsByTypedName(genreOptions, filterState)
                                : genreOptions
                        }
                        getOptionKey={(genreOption) =>
                            typeof genreOption === 'string' ? genreOption : genreOption.key
                        }
                        value={null}
                        inputValue={rename.newName}
                        onInputChange={(_event, typedName, reason) => {
                            if (reason === 'input' || reason === 'clear') {
                                rename.typeNewName(typedName);
                            }
                        }}
                        onChange={(_event, selectedOption) => {
                            if (selectedOption !== null && typeof selectedOption !== 'string') {
                                rename.pickExistingGenre(selectedOption);
                            }
                        }}
                        disabled={rename.isSubmitting || rename.step === 'confirming'}
                        renderInput={(inputProps) => (
                            <TextField
                                {...inputProps}
                                label={t('MUSIC_GENRE_RENAME_FIELD_NAME')}
                                error={rename.nameError !== undefined}
                                helperText={rename.nameError}
                            />
                        )}
                    />
                    {rename.step === 'confirming' && (
                        <RenameConfirmationMessage
                            genreName={genreName}
                            confirmation={rename.confirmation}
                        />
                    )}
                </Stack>
            </DialogContent>
            <DialogActions>
                {rename.step === 'editing' && (
                    <>
                        <Button onClick={onClose} disabled={rename.isSubmitting}>
                            {t('ACTION_CANCEL')}
                        </Button>
                        <Button
                            variant="contained"
                            onClick={rename.review}
                            disabled={!rename.canReview}
                        >
                            {t('MUSIC_GENRE_RENAME_REVIEW')}
                        </Button>
                    </>
                )}
                {rename.step === 'confirming' && (
                    <>
                        <Button onClick={rename.backToEditing} disabled={rename.isSubmitting}>
                            {t('MUSIC_GENRE_RENAME_BACK')}
                        </Button>
                        <Button
                            variant="contained"
                            onClick={rename.confirm}
                            disabled={rename.isSubmitting}
                            startIcon={
                                rename.isSubmitting ? <CircularProgress size={16} /> : undefined
                            }
                        >
                            {t('MUSIC_GENRE_RENAME_CONFIRM')}
                        </Button>
                    </>
                )}
            </DialogActions>
        </Dialog>
    );
}
