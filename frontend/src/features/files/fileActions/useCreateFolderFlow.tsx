import { useState, type ReactNode } from 'react';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import useFile from '@/features/files/providers/fileProvider/fileContext';
import { FileType } from '@/utils';
import { extractBackendErrorMessage } from './bulkOutcome';
import CreateFolderDialog from './createFolderDialog';

export const useCreateFolderFlow = (): {
    openCreateFolderDialog: () => void;
    createFolderDialog: ReactNode;
} => {
    const { selectedItem, createFolder } = useFile();
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const [isOpen, setIsOpen] = useState(false);

    const currentFolderId =
        selectedItem?.type === FileType.Directory ? selectedItem.id : undefined;

    const confirmCreation = async (folderName: string) => {
        try {
            await createFolder(folderName, currentFolderId);
        } catch (error) {
            const errorMessage = extractBackendErrorMessage(error) ?? t('ERROR_CREATE_FOLDER_FAILED');
            enqueueSnackbar(errorMessage, { variant: 'error' });
            return { errorMessage };
        }
        enqueueSnackbar(t('ACTION_CREATE_FOLDER_SUCCESS'), { variant: 'success' });
        setIsOpen(false);
        return { errorMessage: null };
    };

    return {
        openCreateFolderDialog: () => setIsOpen(true),
        createFolderDialog: (
            <CreateFolderDialog
                isOpen={isOpen}
                onClose={() => setIsOpen(false)}
                onConfirm={confirmCreation}
            />
        ),
    };
};

export default useCreateFolderFlow;
