import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import { fileShortcutDefinitions } from './fileShortcutDefinitions';

type KeyboardShortcutsDialogProps = {
    isOpen: boolean;
    onClose: () => void;
};

const KeyboardShortcutsDialog = ({ isOpen, onClose }: KeyboardShortcutsDialogProps) => {
    const { t } = useI18n();

    return (
        <Dialog open={isOpen} onClose={onClose} maxWidth="sm" fullWidth>
            <DialogTitle>{t('FILES_SHORTCUTS_TITLE')}</DialogTitle>
            <DialogContent>
                <Table size="small" aria-label={t('FILES_SHORTCUTS_TITLE')}>
                    <TableHead>
                        <TableRow>
                            <TableCell>{t('FILES_SHORTCUT_COLUMN_KEYS')}</TableCell>
                            <TableCell>{t('FILES_SHORTCUT_COLUMN_ACTION')}</TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {fileShortcutDefinitions.map((shortcut) => (
                            <TableRow key={shortcut.descriptionKey}>
                                <TableCell sx={{ whiteSpace: 'nowrap' }}>
                                    {shortcut.keyLabels.map((keyLabel) => (
                                        <kbd key={keyLabel} style={{ marginRight: 6 }}>
                                            {keyLabel}
                                        </kbd>
                                    ))}
                                </TableCell>
                                <TableCell>{t(shortcut.descriptionKey)}</TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>{t('CLOSE')}</Button>
            </DialogActions>
        </Dialog>
    );
};

export default KeyboardShortcutsDialog;
