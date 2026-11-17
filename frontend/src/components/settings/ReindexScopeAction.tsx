import { useState } from 'react';
import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    LinearProgress,
    Typography,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import useReindexJob from '@/shared/hooks/useReindexJob';
import type { ReindexScope } from '@/service/indexing';

type ReindexScopeActionProps = {
    scope: ReindexScope;
    actionLabelKey: string;
    confirmTitleKey: string;
    confirmMessageKey: string;
};

const ReindexScopeAction = ({
    scope,
    actionLabelKey,
    confirmTitleKey,
    confirmMessageKey,
}: ReindexScopeActionProps) => {
    const { t } = useI18n();
    const [isConfirming, setIsConfirming] = useState(false);
    const { start, isRunning, progress } = useReindexJob();

    const handleConfirm = () => {
        setIsConfirming(false);
        void start(scope);
    };

    return (
        <div>
            <Button variant="outlined" onClick={() => setIsConfirming(true)} disabled={isRunning}>
                {isRunning ? t('SETTINGS_REINDEX_RUNNING') : t(actionLabelKey)}
            </Button>
            {isRunning ? (
                <LinearProgress
                    variant={progress > 0 ? 'determinate' : 'indeterminate'}
                    value={progress}
                    aria-label={t('SETTINGS_REINDEX_RUNNING')}
                />
            ) : null}
            <Dialog
                open={isConfirming}
                onClose={() => setIsConfirming(false)}
                maxWidth="xs"
                fullWidth
            >
                <DialogTitle>{t(confirmTitleKey)}</DialogTitle>
                <DialogContent>
                    <Typography variant="body2">{t(confirmMessageKey)}</Typography>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setIsConfirming(false)}>{t('ACTION_CANCEL')}</Button>
                    <Button onClick={handleConfirm} variant="contained">
                        {t('SETTINGS_REINDEX_CONFIRM')}
                    </Button>
                </DialogActions>
            </Dialog>
        </div>
    );
};

export default ReindexScopeAction;
