import { useRef, useState, type ChangeEvent } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import { Button, Tooltip } from '@mui/material';
import { FolderPlus, Upload } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { documentsQueryKey } from '@/shared/queryKeys/fileQueryKeys';
import { createFolder } from '@/service/files';
import CreateFolderDialog from '@/features/files/fileActions/createFolderDialog';
import { extractBackendErrorMessage } from '@/features/files/fileActions/bulkOutcome';
import { entriesFromFileList } from '@/features/files/upload/entriesFromFileList';
import useUploadQueueContext from '@/features/files/upload/uploadQueueContext';
import useDocumentsDestinationFolder from './useDocumentsDestinationFolder';

interface DocumentsUploadActionsProps {
    selectedFolderId?: number;
}

const hiddenInputStyle = { display: 'none' };

const DocumentsUploadActions = ({ selectedFolderId }: DocumentsUploadActionsProps) => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const queryClient = useQueryClient();
    const { enqueue } = useUploadQueueContext();
    const { status, destinationFolderId, isReady } = useDocumentsDestinationFolder(selectedFolderId);
    const filePickerRef = useRef<HTMLInputElement | null>(null);
    const [isCreateFolderOpen, setIsCreateFolderOpen] = useState(false);

    const disabledReason = status === 'not-configured' ? t('DOCUMENTS_ACTIONS_NOT_CONFIGURED') : '';

    const uploadPickedFiles = (event: ChangeEvent<HTMLInputElement>) => {
        const pickedFiles = event.target.files;
        if (!pickedFiles || pickedFiles.length === 0 || destinationFolderId === undefined) return;
        const entries = entriesFromFileList(pickedFiles);
        event.target.value = '';
        enqueue(entries, destinationFolderId);
    };

    const confirmFolderCreation = async (folderName: string) => {
        try {
            await createFolder(folderName, destinationFolderId);
        } catch (error) {
            const errorMessage = extractBackendErrorMessage(error) ?? t('ERROR_CREATE_FOLDER_FAILED');
            enqueueSnackbar(errorMessage, { variant: 'error' });
            return { errorMessage };
        }
        await queryClient.invalidateQueries({ queryKey: [documentsQueryKey] });
        enqueueSnackbar(t('ACTION_CREATE_FOLDER_SUCCESS'), { variant: 'success' });
        setIsCreateFolderOpen(false);
        return { errorMessage: null };
    };

    return (
        <>
            <Tooltip title={disabledReason}>
                <span>
                    <Button
                        size="small"
                        variant="outlined"
                        startIcon={<Upload size={16} />}
                        disabled={!isReady}
                        onClick={() => filePickerRef.current?.click()}
                    >
                        {t('UPLOAD_FILE')}
                    </Button>
                </span>
            </Tooltip>
            <Tooltip title={disabledReason}>
                <span>
                    <Button
                        size="small"
                        variant="outlined"
                        startIcon={<FolderPlus size={16} />}
                        disabled={!isReady}
                        onClick={() => setIsCreateFolderOpen(true)}
                    >
                        {t('NEW_FOLDER')}
                    </Button>
                </span>
            </Tooltip>
            <input
                ref={filePickerRef}
                type="file"
                multiple
                data-testid="documents-file-picker"
                style={hiddenInputStyle}
                onChange={uploadPickedFiles}
            />
            <CreateFolderDialog
                isOpen={isCreateFolderOpen}
                onClose={() => setIsCreateFolderOpen(false)}
                onConfirm={confirmFolderCreation}
            />
        </>
    );
};

export default DocumentsUploadActions;
