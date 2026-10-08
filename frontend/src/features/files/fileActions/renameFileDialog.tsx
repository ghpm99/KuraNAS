import { FileType } from '@/utils';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import useI18n from '@/components/i18n/provider/i18nContext';
import FileNameDialog, { type FileNameSubmitResult } from './fileNameDialog';
import { findBaseNameLength } from './fileNameValidation';

type RenameFileDialogProps = {
    file: FileData | null;
    onClose: () => void;
    onConfirm: (file: FileData, newName: string) => Promise<FileNameSubmitResult>;
};

const RenameFileDialog = ({ file, onClose, onConfirm }: RenameFileDialogProps) => {
    const { t } = useI18n();
    const currentName = file?.name ?? '';
    const isFolder = file?.type === FileType.Directory;

    const renameFile = async (newName: string): Promise<FileNameSubmitResult> =>
        file ? onConfirm(file, newName) : { errorMessage: null };

    return (
        <FileNameDialog
            isOpen={file !== null}
            title={t('RENAME')}
            confirmLabel={t('RENAME')}
            initialName={currentName}
            initialSelectionEnd={findBaseNameLength(currentName, isFolder)}
            currentName={currentName}
            onClose={onClose}
            onSubmit={renameFile}
        />
    );
};

export default RenameFileDialog;
