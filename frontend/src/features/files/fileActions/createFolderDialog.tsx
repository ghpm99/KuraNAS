import useI18n from '@/components/i18n/provider/i18nContext';
import FileNameDialog, { type FileNameSubmitResult } from './fileNameDialog';

type CreateFolderDialogProps = {
    isOpen: boolean;
    onClose: () => void;
    onConfirm: (folderName: string) => Promise<FileNameSubmitResult>;
};

const CreateFolderDialog = ({ isOpen, onClose, onConfirm }: CreateFolderDialogProps) => {
    const { t } = useI18n();

    return (
        <FileNameDialog
            isOpen={isOpen}
            title={t('NEW_FOLDER')}
            confirmLabel={t('NEW_FOLDER')}
            initialName=""
            initialSelectionEnd={0}
            onClose={onClose}
            onSubmit={onConfirm}
        />
    );
};

export default CreateFolderDialog;
