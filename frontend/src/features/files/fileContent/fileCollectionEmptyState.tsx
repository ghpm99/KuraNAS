import { Button } from '@mui/material';
import { FolderOpen, FolderPlus, Upload } from 'lucide-react';
import EmptyState from '@/components/emptyState/emptyState';
import useI18n from '@/components/i18n/provider/i18nContext';
import useCreateFolderFlow from '@/features/files/fileActions/useCreateFolderFlow';
import type { FileListCategoryType } from '@/features/files/providers/fileProvider/fileContext';
import useUploadPickers from '@/features/files/upload/useUploadPickers';

const EmptyFolderState = () => {
    const { t } = useI18n();
    const { openFilePicker, pickerInputs } = useUploadPickers();
    const { openCreateFolderDialog, createFolderDialog } = useCreateFolderFlow();

    return (
        <EmptyState
            icon={<FolderOpen size={32} />}
            title={t('FILES_EMPTY_FOLDER_TITLE')}
            description={t('FILES_EMPTY_FOLDER_DESCRIPTION')}
            actions={
                <>
                    {pickerInputs}
                    <Button
                        variant="contained"
                        size="small"
                        startIcon={<Upload size={16} />}
                        onClick={openFilePicker}
                    >
                        {t('UPLOAD_FILE')}
                    </Button>
                    <Button
                        variant="outlined"
                        size="small"
                        startIcon={<FolderPlus size={16} />}
                        onClick={openCreateFolderDialog}
                    >
                        {t('NEW_FOLDER')}
                    </Button>
                    {createFolderDialog}
                </>
            }
        />
    );
};

type FileCollectionEmptyStateProps = {
    customMessage?: string;
    isFolderOpen: boolean;
    fileListFilter: FileListCategoryType;
};

const FileCollectionEmptyState = ({
    customMessage,
    isFolderOpen,
    fileListFilter,
}: FileCollectionEmptyStateProps) => {
    const { t } = useI18n();

    if (customMessage) return <EmptyState title={customMessage} />;
    if (isFolderOpen) return <EmptyFolderState />;
    if (fileListFilter === 'starred') {
        return (
            <EmptyState
                title={t('FILES_EMPTY_FAVORITES_TITLE')}
                description={t('FILES_EMPTY_FAVORITES_DESCRIPTION')}
            />
        );
    }
    if (fileListFilter === 'recent') {
        return (
            <EmptyState
                title={t('FILES_EMPTY_RECENT_TITLE')}
                description={t('FILES_EMPTY_RECENT_DESCRIPTION')}
            />
        );
    }
    return <EmptyState title={t('EMPTY_FILE_LIST')} />;
};

export default FileCollectionEmptyState;
