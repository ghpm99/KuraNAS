import {
    Alert,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    Stack,
    TextField,
} from '@mui/material';
import { useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useVideoEpisodeCorrection } from '@/features/videos/providers/videoContentProvider/useVideoEpisodeCorrection';

type VideoEpisodeDialogProps = {
    isOpen: boolean;
    fileId: number;
    fileName: string;
    initialSeason?: number | null;
    initialEpisode?: number | null;
    onClose: () => void;
};

const toInputText = (number?: number | null) => (number == null ? '' : String(number));

const parseInputNumber = (text: string): number | null => {
    if (text.trim() === '') return null;
    const parsedNumber = Number(text);
    return Number.isNaN(parsedNumber) ? null : parsedNumber;
};

export default function VideoEpisodeDialog({
    isOpen,
    fileId,
    fileName,
    initialSeason,
    initialEpisode,
    onClose,
}: VideoEpisodeDialogProps) {
    const { t } = useI18n();
    const [seasonText, setSeasonText] = useState(toInputText(initialSeason));
    const [episodeText, setEpisodeText] = useState(toInputText(initialEpisode));
    const { correct, isCorrecting, errorMessage, errorField, resetError } =
        useVideoEpisodeCorrection(fileId, onClose);

    const submit = () =>
        correct({
            season: parseInputNumber(seasonText),
            episode: parseInputNumber(episodeText),
        });

    return (
        <Dialog open={isOpen} onClose={onClose} maxWidth="xs" fullWidth>
            <DialogTitle>{t('VIDEO_CORRECT_EPISODE')}</DialogTitle>
            <DialogContent>
                <p>{fileName}</p>
                {errorMessage && <Alert severity="error">{errorMessage}</Alert>}
                <Stack direction="row" spacing={2} sx={{ mt: 1 }}>
                    <TextField
                        type="number"
                        label={t('VIDEO_CORRECTION_SEASON_LABEL')}
                        value={seasonText}
                        error={errorField === 'season'}
                        slotProps={{ htmlInput: { min: 0 } }}
                        onChange={(event) => {
                            resetError();
                            setSeasonText(event.target.value);
                        }}
                    />
                    <TextField
                        type="number"
                        label={t('VIDEO_CORRECTION_EPISODE_LABEL')}
                        value={episodeText}
                        error={errorField === 'episode'}
                        slotProps={{ htmlInput: { min: 1 } }}
                        onChange={(event) => {
                            resetError();
                            setEpisodeText(event.target.value);
                        }}
                    />
                </Stack>
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>{t('ACTION_CANCEL')}</Button>
                <Button
                    disabled={isCorrecting}
                    onClick={() => correct({ season: null, episode: null })}
                >
                    {t('VIDEO_CORRECTION_AUTOMATIC')}
                </Button>
                <Button variant="contained" disabled={isCorrecting} onClick={submit}>
                    {t('VIDEO_CORRECTION_SAVE')}
                </Button>
            </DialogActions>
        </Dialog>
    );
}
