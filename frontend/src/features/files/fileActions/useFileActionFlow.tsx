import { useCallback, useEffect, useRef, useState, type ReactNode } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import FolderPicker, { type FolderPickerResult } from '@/components/folderPicker/folderPicker';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import { focusFileItem } from '@/features/files/shortcuts/fileItemFocus';
import { useFileSelectionContext } from '@/features/files/selection/fileSelectionContext';
import type { BulkOutcome } from './bulkOutcome';
import DeleteItemsDialog from '@/components/deleteItemsDialog/deleteItemsDialog';
import RenameFileDialog from './renameFileDialog';
import useFileOperations from './useFileOperations';

export type DialogFileAction = 'move' | 'copy' | 'rename' | 'delete';

type PendingAction = { action: DialogFileAction; files: FileData[] };

export const useFileActionFlow = (): {
    startAction: (action: DialogFileAction, files: FileData[]) => void;
    dialogs: ReactNode;
} => {
    const operations = useFileOperations();
    const { t } = useI18n();
    const { deselect } = useFileSelectionContext();
    const [pendingAction, setPendingAction] = useState<PendingAction | null>(null);

    const focusReturnTargetIdRef = useRef<number | null>(null);

    const startAction = useCallback((action: DialogFileAction, files: FileData[]) => {
        const [firstFile] = files;
        if (!firstFile) return;
        focusReturnTargetIdRef.current = firstFile.id;
        setPendingAction({ action, files });
    }, []);

    const isActionPending = pendingAction !== null;
    useEffect(() => {
        if (isActionPending || focusReturnTargetIdRef.current === null) return;
        focusFileItem(focusReturnTargetIdRef.current);
        focusReturnTargetIdRef.current = null;
    }, [isActionPending]);

    const cancelAction = () => setPendingAction(null);

    const finishAction = (action: DialogFileAction, outcome: BulkOutcome) => {
        deselect(outcome.succeededFiles);
        setPendingAction(
            outcome.failedFiles.length === 0 ? null : { action, files: outcome.failedFiles }
        );
    };

    const pendingFiles = pendingAction?.files ?? [];
    const isPending = (action: DialogFileAction) => pendingAction?.action === action;

    const applyDestination =
        (
            action: 'move' | 'copy',
            run: (files: FileData[], destination: FolderPickerResult) => Promise<BulkOutcome>
        ) =>
        async (destination: FolderPickerResult) => {
            finishAction(action, await run(pendingFiles, destination));
        };

    const renameTarget = isPending('rename') ? (pendingFiles[0] ?? null) : null;

    const dialogs = (
        <>
            <FolderPicker
                open={isPending('move')}
                mode="move"
                onClose={cancelAction}
                onSelect={applyDestination('move', operations.moveFiles)}
            />
            <FolderPicker
                open={isPending('copy')}
                mode="copy"
                onClose={cancelAction}
                onSelect={applyDestination('copy', operations.copyFiles)}
            />
            <RenameFileDialog
                key={`${renameTarget?.id}:${renameTarget?.name}`}
                file={renameTarget}
                onClose={cancelAction}
                onConfirm={async (file, newName) => {
                    const outcome = await operations.renameSingleFile(file, newName);
                    finishAction('rename', outcome);
                    if (outcome.failedFiles.length === 0) return { errorMessage: null };
                    return { errorMessage: outcome.firstFailureMessage ?? t('ERROR_RENAME_FAILED') };
                }}
            />
            <DeleteItemsDialog
                items={pendingFiles}
                isOpen={isPending('delete')}
                onClose={cancelAction}
                onConfirm={async (files, isPermanent) =>
                    finishAction('delete', await operations.deleteFiles(files, isPermanent))
                }
            />
        </>
    );

    return { startAction, dialogs };
};

export default useFileActionFlow;
