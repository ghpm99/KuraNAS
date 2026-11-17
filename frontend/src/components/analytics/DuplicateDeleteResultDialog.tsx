import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    Typography,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { DuplicateCopiesDeletionOutcome } from '@/components/analytics/useDuplicateCopiesDeletion';

type DuplicateDeleteResultDialogProps = {
    outcome: DuplicateCopiesDeletionOutcome | null;
    onClose: () => void;
};

const DuplicateDeleteResultDialog = ({ outcome, onClose }: DuplicateDeleteResultDialogProps) => {
    const { t } = useI18n();

    return (
        <Dialog open={outcome !== null} onClose={onClose} maxWidth="sm" fullWidth>
            <DialogTitle>{t('ANALYTICS_DUPLICATES_RESULT_TITLE')}</DialogTitle>
            <DialogContent>
                <Typography variant="body2" gutterBottom>
                    {t('ANALYTICS_DUPLICATES_RESULT_TRASHED', {
                        count: String(outcome?.trashedCount ?? 0),
                    })}
                </Typography>
                {outcome && outcome.failures.length > 0 ? (
                    <>
                        <Typography variant="body2" color="error" gutterBottom>
                            {t('ANALYTICS_DUPLICATES_RESULT_FAILED', {
                                count: String(outcome.failures.length),
                            })}
                        </Typography>
                        <ul>
                            {outcome.failures.map(({ copy, message }) => (
                                <li key={copy.id}>
                                    {t('ANALYTICS_DUPLICATES_RESULT_FAILURE_ITEM', {
                                        path: copy.path,
                                        message,
                                    })}
                                </li>
                            ))}
                        </ul>
                    </>
                ) : null}
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>{t('CLOSE')}</Button>
            </DialogActions>
        </Dialog>
    );
};

export default DuplicateDeleteResultDialog;
