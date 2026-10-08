import { useCallback, useState } from 'react';
import { Button, Dialog, DialogActions, DialogContent, DialogTitle, TextField } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import { fileNameIssueMessageKeys, findFileNameIssue } from './fileNameValidation';

export type FileNameSubmitResult = { errorMessage: string | null };

type FileNameDialogProps = {
    isOpen: boolean;
    title: string;
    confirmLabel: string;
    initialName: string;
    initialSelectionEnd: number;
    currentName?: string;
    onClose: () => void;
    onSubmit: (name: string) => Promise<FileNameSubmitResult>;
};

type FileNameDialogBodyProps = Omit<FileNameDialogProps, 'isOpen'>;

const FileNameDialogBody = ({
    title,
    confirmLabel,
    initialName,
    initialSelectionEnd,
    currentName,
    onClose,
    onSubmit,
}: FileNameDialogBodyProps) => {
    const { t } = useI18n();
    const [typedName, setTypedName] = useState(initialName);
    const [hasTyped, setHasTyped] = useState(false);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [backendErrorMessage, setBackendErrorMessage] = useState<string | null>(null);

    const focusAndSelectBaseName = useCallback(
        (input: HTMLInputElement | null) => {
            if (!input) return;
            input.focus();
            input.setSelectionRange(0, initialSelectionEnd);
        },
        [initialSelectionEnd]
    );

    const nameIssue = findFileNameIssue(typedName, currentName);
    const isIssueVisible = nameIssue !== null && (hasTyped || nameIssue === 'unchanged');
    const isIssueBlocking = nameIssue !== null && nameIssue !== 'unchanged';
    const helperText = backendErrorMessage ?? (isIssueVisible ? t(fileNameIssueMessageKeys[nameIssue]) : ' ');
    const hasError = backendErrorMessage !== null || (isIssueVisible && isIssueBlocking);

    const changeName = (nextName: string) => {
        setHasTyped(true);
        setTypedName(nextName);
        setBackendErrorMessage(null);
    };

    const submitName = async () => {
        if (nameIssue !== null || isSubmitting) return;
        setIsSubmitting(true);
        const { errorMessage } = await onSubmit(typedName.trim());
        if (errorMessage === null) return;
        setBackendErrorMessage(errorMessage);
        setIsSubmitting(false);
    };

    return (
        <form
            onSubmit={(event) => {
                event.preventDefault();
                void submitName();
            }}
        >
            <DialogTitle>{title}</DialogTitle>
            <DialogContent>
                <TextField
                    margin="dense"
                    label={t('NAME')}
                    fullWidth
                    value={typedName}
                    onChange={(event) => changeName(event.target.value)}
                    error={hasError}
                    helperText={helperText}
                    inputRef={focusAndSelectBaseName}
                    slotProps={{ formHelperText: { role: hasError ? 'alert' : undefined } }}
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>{t('ACTION_CANCEL')}</Button>
                <Button type="submit" variant="contained" disabled={nameIssue !== null || isSubmitting}>
                    {confirmLabel}
                </Button>
            </DialogActions>
        </form>
    );
};

const FileNameDialog = ({ isOpen, onClose, ...bodyProps }: FileNameDialogProps) => (
    <Dialog open={isOpen} onClose={onClose} maxWidth="sm" fullWidth>
        <FileNameDialogBody {...bodyProps} onClose={onClose} />
    </Dialog>
);

export default FileNameDialog;
