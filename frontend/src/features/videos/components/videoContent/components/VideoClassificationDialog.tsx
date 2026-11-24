import {
    Alert,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    FormControl,
    FormControlLabel,
    FormLabel,
    Radio,
    RadioGroup,
} from '@mui/material';
import { useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useVideoClassificationCorrection } from '@/features/videos/providers/videoContentProvider/useVideoClassificationCorrection';
import {
    videoManualClassifications,
    type VideoManualClassification,
} from '@/service/videoPlayback';
import { getVideoClassificationLabelKey } from '../videoClassificationLabel';

const automaticOption = 'automatic';

type VideoClassificationDialogProps = {
    isOpen: boolean;
    fileId: number;
    fileName: string;
    onClose: () => void;
};

export default function VideoClassificationDialog({
    isOpen,
    fileId,
    fileName,
    onClose,
}: VideoClassificationDialogProps) {
    const { t } = useI18n();
    const [selectedOption, setSelectedOption] = useState<string>(automaticOption);
    const { correct, isCorrecting, errorMessage, errorField, resetError } =
        useVideoClassificationCorrection(fileId, onClose);
    const isClassificationFieldInvalid = errorField === 'classification';

    const submit = () =>
        correct(
            selectedOption === automaticOption
                ? null
                : (selectedOption as VideoManualClassification)
        );

    return (
        <Dialog open={isOpen} onClose={onClose} maxWidth="xs" fullWidth>
            <DialogTitle>{t('VIDEO_CORRECT_CLASSIFICATION')}</DialogTitle>
            <DialogContent>
                <p>{fileName}</p>
                {errorMessage && <Alert severity="error">{errorMessage}</Alert>}
                <FormControl error={isClassificationFieldInvalid}>
                    <FormLabel>{t('VIDEO_CORRECTION_CLASSIFICATION_LABEL')}</FormLabel>
                    <RadioGroup
                        value={selectedOption}
                        onChange={(event) => {
                            resetError();
                            setSelectedOption(event.target.value);
                        }}
                    >
                        <FormControlLabel
                            value={automaticOption}
                            control={<Radio />}
                            label={t('VIDEO_CORRECTION_AUTOMATIC')}
                        />
                        {videoManualClassifications.map((classification) => (
                            <FormControlLabel
                                key={classification}
                                value={classification}
                                control={<Radio />}
                                label={t(
                                    classification === 'other'
                                        ? 'VIDEO_CLASS_OTHER'
                                        : getVideoClassificationLabelKey(classification)
                                )}
                            />
                        ))}
                    </RadioGroup>
                </FormControl>
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>{t('ACTION_CANCEL')}</Button>
                <Button variant="contained" disabled={isCorrecting} onClick={submit}>
                    {t('VIDEO_CORRECTION_SAVE')}
                </Button>
            </DialogActions>
        </Dialog>
    );
}
