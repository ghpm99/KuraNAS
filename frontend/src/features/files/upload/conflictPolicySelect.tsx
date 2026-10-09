import { MenuItem, TextField } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { UploadConflictPolicy } from '@/service/files';
import useUploadQueueContext from './uploadQueueContext';

const policyLabelKeys: Record<UploadConflictPolicy, string> = {
    rename: 'FILES_UPLOAD_ON_CONFLICT_RENAME',
    replace: 'FILES_UPLOAD_ON_CONFLICT_REPLACE',
    skip: 'FILES_UPLOAD_ON_CONFLICT_SKIP',
};

const policyOrder: UploadConflictPolicy[] = ['rename', 'replace', 'skip'];

const ConflictPolicySelect = () => {
    const { t } = useI18n();
    const { conflictPolicy, setConflictPolicy } = useUploadQueueContext();

    return (
        <TextField
            select
            size="small"
            label={t('FILES_UPLOAD_ON_CONFLICT_LABEL')}
            value={conflictPolicy}
            onChange={(event) => setConflictPolicy(event.target.value as UploadConflictPolicy)}
            sx={{ minWidth: 140 }}
        >
            {policyOrder.map((policy) => (
                <MenuItem key={policy} value={policy}>
                    {t(policyLabelKeys[policy])}
                </MenuItem>
            ))}
        </TextField>
    );
};

export default ConflictPolicySelect;
