import { TextField } from '@mui/material';
import styles from './TrashScreen.module.css';
import { useTrashRetention } from './useTrashRetention';

const TrashRetentionField = () => {
    const {
        t,
        draftDays,
        changeDraft,
        commitDraft,
        isDraftInvalid,
        isRetentionLoading,
        isRetentionUnknown,
        isSaving,
    } = useTrashRetention();

    const helperText = isDraftInvalid
        ? t('TRASH_RETENTION_INVALID')
        : isRetentionUnknown
          ? t('TRASH_RETENTION_UNKNOWN')
          : undefined;

    return (
        <div className={styles.retention}>
            <TextField
                size="small"
                type="number"
                label={t('TRASH_RETENTION_LABEL')}
                value={draftDays}
                disabled={isRetentionLoading || isSaving}
                error={isDraftInvalid}
                helperText={helperText}
                slotProps={{ htmlInput: { min: 1, step: 1 } }}
                onChange={(event) => changeDraft(event.target.value)}
                onBlur={() => void commitDraft()}
            />
            <span className={styles.description}>{t('TRASH_RETENTION_HINT')}</span>
        </div>
    );
};

export default TrashRetentionField;
