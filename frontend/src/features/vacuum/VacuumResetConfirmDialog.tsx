import { Button, Dialog, DialogActions, DialogContent, DialogTitle, Typography } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';

type VacuumResetConfirmDialogProps = {
    partLabel: string | undefined;
    isResetting: boolean;
    onCancel: () => void;
    onConfirm: () => void;
};

const VacuumResetConfirmDialog = ({
    partLabel,
    isResetting,
    onCancel,
    onConfirm,
}: VacuumResetConfirmDialogProps) => {
    const { t } = useI18n();

    return (
        <Dialog open={partLabel !== undefined} onClose={onCancel} maxWidth="xs" fullWidth>
            <DialogTitle>{t('VACUUM_RESET_CONFIRM_TITLE', { part: partLabel ?? '' })}</DialogTitle>
            <DialogContent>
                <Typography variant="body2">{t('VACUUM_RESET_CONFIRM_MESSAGE')}</Typography>
            </DialogContent>
            <DialogActions>
                <Button onClick={onCancel} disabled={isResetting}>
                    {t('ACTION_CANCEL')}
                </Button>
                <Button onClick={onConfirm} disabled={isResetting} variant="contained" color="error">
                    {t('VACUUM_RESET_ACTION')}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default VacuumResetConfirmDialog;
