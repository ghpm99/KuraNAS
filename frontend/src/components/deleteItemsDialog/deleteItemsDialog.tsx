import { useState } from 'react';
import { Link as RouterLink, useInRouterContext } from 'react-router-dom';
import {
    Alert,
    Button,
    Checkbox,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    FormControlLabel,
    Link,
    Typography,
} from '@mui/material';
import { appRoutes } from '@/app/routes';
import useI18n from '@/components/i18n/provider/i18nContext';

type DeleteItemsDialogProps<TItem> = {
    items: TItem[];
    isOpen: boolean;
    onClose: () => void;
    onConfirm: (items: TItem[], isPermanent: boolean) => void;
};

const TrashPageLink = ({ onNavigate }: { onNavigate: () => void }) => {
    const { t } = useI18n();
    const isInsideRouter = useInRouterContext();
    const label = t('FILES_DELETE_OPEN_TRASH');

    if (!isInsideRouter) return <Link href={appRoutes.trash}>{label}</Link>;
    return (
        <Link component={RouterLink} to={appRoutes.trash} onClick={onNavigate}>
            {label}
        </Link>
    );
};

type DeleteItemsDialogBodyProps<TItem> = Omit<DeleteItemsDialogProps<TItem>, 'isOpen'>;

const DeleteItemsDialogBody = <TItem,>({
    items,
    onClose,
    onConfirm,
}: DeleteItemsDialogBodyProps<TItem>) => {
    const { t } = useI18n();
    const [isPermanent, setIsPermanent] = useState(false);
    const confirmationMessage =
        items.length > 1
            ? t('FILES_CONFIRM_DELETE_MANY', { count: String(items.length) })
            : t('CONFIRM_DELETE');

    return (
        <>
            <DialogTitle>{t('DELETE')}</DialogTitle>
            <DialogContent>
                <Typography variant="body2" gutterBottom>
                    {confirmationMessage}
                </Typography>
                <Typography variant="body2" color="text.secondary" gutterBottom>
                    {t('FILES_DELETE_TRASH_NOTICE')}{' '}
                    <TrashPageLink onNavigate={onClose} />
                </Typography>
                <FormControlLabel
                    label={t('FILES_DELETE_PERMANENTLY')}
                    control={
                        <Checkbox
                            checked={isPermanent}
                            onChange={(event) => setIsPermanent(event.target.checked)}
                        />
                    }
                />
                {isPermanent ? (
                    <Alert severity="error" sx={{ mt: 1 }}>
                        {t('FILES_DELETE_PERMANENT_WARNING')}
                    </Alert>
                ) : null}
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>{t('ACTION_CANCEL')}</Button>
                <Button onClick={() => onConfirm(items, isPermanent)} variant="contained" color="error">
                    {isPermanent ? t('FILES_DELETE_PERMANENTLY') : t('DELETE')}
                </Button>
            </DialogActions>
        </>
    );
};

const DeleteItemsDialog = <TItem,>({
    isOpen,
    onClose,
    ...bodyProps
}: DeleteItemsDialogProps<TItem>) => (
    <Dialog open={isOpen} onClose={onClose} maxWidth="xs" fullWidth>
        <DeleteItemsDialogBody {...bodyProps} onClose={onClose} />
    </Dialog>
);

export default DeleteItemsDialog;
