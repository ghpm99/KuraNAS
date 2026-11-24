import {
    Button,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    TextField,
    ToggleButton,
    ToggleButtonGroup,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import { PlaylistKindChoice } from '@/features/music/providers/playlistsProvider/playlistsContext';
import { hasPlaylistRuleCriteria } from '@/features/music/providers/playlistsProvider/playlistRuleCriteria';
import { PlaylistRules } from '@/types/playlist';
import PlaylistRuleForm from './PlaylistRuleForm';

type PlaylistCreateDialogProps = {
    open: boolean;
    newName: string;
    newDescription: string;
    newKind: PlaylistKindChoice;
    newRules: PlaylistRules;
    errorField: string | null;
    isSubmitting: boolean;
    onKindChange: (kind: PlaylistKindChoice) => void;
    onRulesChange: (rules: PlaylistRules) => void;
    onClose: () => void;
    onNameChange: (value: string) => void;
    onDescriptionChange: (value: string) => void;
    onSubmit: () => void;
};

export default function PlaylistCreateDialog({
    open,
    newName,
    newDescription,
    newKind,
    newRules,
    errorField,
    isSubmitting,
    onKindChange,
    onRulesChange,
    onClose,
    onNameChange,
    onDescriptionChange,
    onSubmit,
}: PlaylistCreateDialogProps) {
    const { t } = useI18n();
    const isRule = newKind === 'rule';
    const isSubmitBlocked =
        !newName.trim() || isSubmitting || (isRule && !hasPlaylistRuleCriteria(newRules));

    return (
        <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
            <DialogTitle>{t('MUSIC_CREATE_PLAYLIST')}</DialogTitle>
            <DialogContent>
                <ToggleButtonGroup
                    exclusive
                    size="small"
                    aria-label={t('MUSIC_PLAYLIST_KIND_LABEL')}
                    value={newKind}
                    onChange={(_event, nextKind) => nextKind && onKindChange(nextKind)}
                    sx={{ mt: 1 }}
                >
                    <ToggleButton value="manual">{t('MUSIC_PLAYLIST_KIND_MANUAL')}</ToggleButton>
                    <ToggleButton value="rule">{t('MUSIC_PLAYLIST_KIND_RULE')}</ToggleButton>
                </ToggleButtonGroup>
                <TextField
                    autoFocus
                    fullWidth
                    label={t('NAME')}
                    value={newName}
                    onChange={(event) => onNameChange(event.target.value)}
                    sx={{ mt: 2, mb: 2 }}
                />
                <TextField
                    fullWidth
                    label={t('MUSIC_DESCRIPTION_OPTIONAL')}
                    value={newDescription}
                    onChange={(event) => onDescriptionChange(event.target.value)}
                    multiline
                    rows={2}
                />
                {isRule && (
                    <PlaylistRuleForm
                        rules={newRules}
                        errorField={errorField}
                        onChange={onRulesChange}
                    />
                )}
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>{t('ACTION_CANCEL')}</Button>
                <Button variant="contained" onClick={onSubmit} disabled={isSubmitBlocked}>
                    {isSubmitting ? <CircularProgress size={20} /> : t('ACTION_CREATE')}
                </Button>
            </DialogActions>
        </Dialog>
    );
}
