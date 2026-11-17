import { useState } from 'react';
import {
    Button,
    Checkbox,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    FormControlLabel,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import useUploadQueueContext from './uploadQueueContext';
import {
    isAwaitingDecision,
    type ResolvedConflictChoice,
    type UploadItem,
} from './uploadQueueTypes';

const resolutionButtons: { choice: ResolvedConflictChoice; labelKey: string }[] = [
    { choice: 'rename', labelKey: 'FILES_UPLOAD_ON_CONFLICT_RENAME' },
    { choice: 'replace', labelKey: 'FILES_UPLOAD_ON_CONFLICT_REPLACE' },
    { choice: 'skip', labelKey: 'FILES_UPLOAD_ON_CONFLICT_SKIP' },
];

const countItemsCoveredByApplyToAll = (items: UploadItem[]): number =>
    items.filter(
        (item) =>
            isAwaitingDecision(item) || (item.status === 'queued' && item.onConflict === 'ask')
    ).length;

const UploadConflictDialog = () => {
    const { t } = useI18n();
    const { items, resolveConflict, dismissConflict } = useUploadQueueContext();
    const [isApplyToAllChecked, setIsApplyToAllChecked] = useState(false);

    const conflictingItem = items.find(isAwaitingDecision);
    if (!conflictingItem) return null;

    const coveredCount = countItemsCoveredByApplyToAll(items);
    const canApplyToAll = coveredCount > 1;

    const resolve = (choice: ResolvedConflictChoice) => {
        resolveConflict(conflictingItem.id, choice, canApplyToAll && isApplyToAllChecked);
        setIsApplyToAllChecked(false);
    };

    const dismiss = () => {
        dismissConflict(conflictingItem.id);
        setIsApplyToAllChecked(false);
    };

    return (
        <Dialog open onClose={dismiss} aria-labelledby="upload-conflict-title">
            <DialogTitle id="upload-conflict-title">{t('FILES_UPLOAD_CONFLICT_TITLE')}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    {t('FILES_UPLOAD_CONFLICT_MESSAGE', { name: conflictingItem.displayName })}
                </DialogContentText>
                {canApplyToAll ? (
                    <FormControlLabel
                        control={
                            <Checkbox
                                checked={isApplyToAllChecked}
                                onChange={(event) => setIsApplyToAllChecked(event.target.checked)}
                            />
                        }
                        label={t('FILES_UPLOAD_CONFLICT_APPLY_TO_ALL', {
                            count: String(coveredCount),
                        })}
                    />
                ) : null}
            </DialogContent>
            <DialogActions>
                <Button color="inherit" onClick={dismiss}>
                    {t('FILES_UPLOAD_CONFLICT_CANCEL')}
                </Button>
                {resolutionButtons.map(({ choice, labelKey }) => (
                    <Button key={choice} onClick={() => resolve(choice)}>
                        {t(labelKey)}
                    </Button>
                ))}
            </DialogActions>
        </Dialog>
    );
};

export default UploadConflictDialog;
