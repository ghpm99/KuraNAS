import { useState } from 'react';
import {
    Button,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    TextField,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import {
    hasPlaylistRuleCriteria,
    normalizePlaylistRules,
} from '@/features/music/providers/playlistsProvider/playlistRuleCriteria';
import { PlaylistRules } from '@/types/playlist';
import PlaylistRuleForm from './PlaylistRuleForm';

type PlaylistEditDialogProps = {
    currentName: string;
    currentDescription: string;
    currentRules?: PlaylistRules | null;
    errorField?: string | null;
    isSubmitting: boolean;
    onClose: () => void;
    onSubmit: (name: string, description: string, rules?: PlaylistRules) => void;
};

export default function PlaylistEditDialog({
    currentName,
    currentDescription,
    currentRules,
    errorField,
    isSubmitting,
    onClose,
    onSubmit,
}: PlaylistEditDialogProps) {
    const { t } = useI18n();
    const [draftName, setDraftName] = useState(currentName);
    const [draftDescription, setDraftDescription] = useState(currentDescription);
    const [draftRules, setDraftRules] = useState<PlaylistRules | null>(() =>
        currentRules ? normalizePlaylistRules(currentRules) : null
    );
    const isRuleMissingCriteria = draftRules !== null && !hasPlaylistRuleCriteria(draftRules);

    return (
        <Dialog open onClose={onClose} maxWidth="sm" fullWidth>
            <DialogTitle>{t('MUSIC_PLAYLIST_EDIT')}</DialogTitle>
            <DialogContent>
                <TextField
                    autoFocus
                    fullWidth
                    label={t('NAME')}
                    value={draftName}
                    onChange={(event) => setDraftName(event.target.value)}
                    sx={{ mt: 1, mb: 2 }}
                />
                <TextField
                    fullWidth
                    label={t('MUSIC_DESCRIPTION_OPTIONAL')}
                    value={draftDescription}
                    onChange={(event) => setDraftDescription(event.target.value)}
                    multiline
                    rows={2}
                />
                {draftRules && (
                    <PlaylistRuleForm
                        rules={draftRules}
                        errorField={errorField}
                        onChange={setDraftRules}
                    />
                )}
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>{t('ACTION_CANCEL')}</Button>
                <Button
                    variant="contained"
                    onClick={() =>
                        draftRules
                            ? onSubmit(draftName.trim(), draftDescription.trim(), draftRules)
                            : onSubmit(draftName.trim(), draftDescription.trim())
                    }
                    disabled={!draftName.trim() || isSubmitting || isRuleMissingCriteria}
                >
                    {isSubmitting ? <CircularProgress size={20} /> : t('MUSIC_PLAYLIST_SAVE')}
                </Button>
            </DialogActions>
        </Dialog>
    );
}
