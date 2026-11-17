import { useState } from 'react';
import {
    Alert,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    Typography,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';

type DuplicateDeleteDialogProps = {
    copiesCount: number;
    groupsLeftWithoutCopiesCount: number;
    isDeleting: boolean;
    onCancel: () => void;
    onConfirm: () => void;
};

type ConfirmationStep = 'regular' | 'emptiedGroups';

const DuplicateDeleteDialog = ({
    copiesCount,
    groupsLeftWithoutCopiesCount,
    isDeleting,
    onCancel,
    onConfirm,
}: DuplicateDeleteDialogProps) => {
    const { t } = useI18n();
    const [confirmationStep, setConfirmationStep] = useState<ConfirmationStep>('regular');
    const isOpen = copiesCount > 0;
    const needsExtraConfirmation =
        groupsLeftWithoutCopiesCount > 0 && confirmationStep === 'regular';

    const cancel = () => {
        setConfirmationStep('regular');
        onCancel();
    };

    const confirm = () => {
        if (needsExtraConfirmation) {
            setConfirmationStep('emptiedGroups');
            return;
        }
        setConfirmationStep('regular');
        onConfirm();
    };

    const message =
        copiesCount > 1
            ? t('FILES_CONFIRM_DELETE_MANY', { count: String(copiesCount) })
            : t('CONFIRM_DELETE');

    return (
        <Dialog open={isOpen} onClose={isDeleting ? undefined : cancel} maxWidth="xs" fullWidth>
            <DialogTitle>{t('DELETE')}</DialogTitle>
            <DialogContent>
                {confirmationStep === 'emptiedGroups' ? (
                    <Alert severity="error">
                        {t('ANALYTICS_DUPLICATES_DELETE_ALL_COPIES_WARNING', {
                            count: String(groupsLeftWithoutCopiesCount),
                        })}
                    </Alert>
                ) : (
                    <>
                        <Typography variant="body2" gutterBottom>
                            {message}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                            {t('FILES_DELETE_TRASH_NOTICE')}
                        </Typography>
                    </>
                )}
            </DialogContent>
            <DialogActions>
                <Button onClick={cancel} disabled={isDeleting}>
                    {t('ACTION_CANCEL')}
                </Button>
                <Button onClick={confirm} disabled={isDeleting} variant="contained" color="error">
                    {confirmationStep === 'emptiedGroups'
                        ? t('ANALYTICS_DUPLICATES_DELETE_ANYWAY')
                        : t('DELETE')}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default DuplicateDeleteDialog;
