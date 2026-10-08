import { useCallback } from 'react';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import useFile, { type FileData } from '@/features/files/providers/fileProvider/fileContext';
import { getFileDownloadUrl, getFilesZipDownloadUrl } from '@/service/files';
import { triggerBrowserDownload } from '@/service/browserDownload';
import type { FolderPickerResult } from '@/components/folderPicker/folderPicker';
import { runBulkOperation, type BulkOutcome } from './bulkOutcome';

type OutcomeMessages = {
    singleSuccessKey: string;
    multipleSuccessKey: string;
    failureKey: string;
};

const moveMessages: OutcomeMessages = {
    singleSuccessKey: 'ACTION_MOVE_SUCCESS',
    multipleSuccessKey: 'FILES_BULK_MOVE_SUCCESS',
    failureKey: 'ERROR_MOVE_FAILED',
};

const copyMessages: OutcomeMessages = {
    singleSuccessKey: 'ACTION_COPY_SUCCESS',
    multipleSuccessKey: 'FILES_BULK_COPY_SUCCESS',
    failureKey: 'ERROR_COPY_FAILED',
};

const deleteMessages: OutcomeMessages = {
    singleSuccessKey: 'ACTION_DELETE_SUCCESS',
    multipleSuccessKey: 'FILES_BULK_DELETE_SUCCESS',
    failureKey: 'ERROR_DELETE_FAILED',
};

const favoriteMessages: OutcomeMessages = {
    singleSuccessKey: 'FILES_FAVORITE_SUCCESS',
    multipleSuccessKey: 'FILES_BULK_FAVORITE_SUCCESS',
    failureKey: 'ERROR_FAVORITE_FAILED',
};

const renameMessages: OutcomeMessages = {
    singleSuccessKey: 'ACTION_RENAME_SUCCESS',
    multipleSuccessKey: 'ACTION_RENAME_SUCCESS',
    failureKey: 'ERROR_RENAME_FAILED',
};

export const useFileOperations = () => {
    const { moveFile, copyFile, deleteFile, renameFile, toggleStarred } = useFile();
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();

    const notifyOutcome = useCallback(
        (outcome: BulkOutcome, messages: OutcomeMessages) => {
            const succeededCount = outcome.succeededFiles.length;
            const failedCount = outcome.failedFiles.length;
            const failureMessage = outcome.firstFailureMessage ?? t(messages.failureKey);

            if (failedCount === 0) {
                enqueueSnackbar(
                    succeededCount === 1
                        ? t(messages.singleSuccessKey)
                        : t(messages.multipleSuccessKey, { count: String(succeededCount) }),
                    { variant: 'success' }
                );
                return;
            }
            if (succeededCount === 0 && failedCount === 1) {
                enqueueSnackbar(failureMessage, { variant: 'error' });
                return;
            }
            if (succeededCount === 0) {
                enqueueSnackbar(
                    t('FILES_BULK_FAILED_SUMMARY', { failed: String(failedCount), message: failureMessage }),
                    { variant: 'error' }
                );
                return;
            }
            enqueueSnackbar(
                t('FILES_BULK_PARTIAL_SUMMARY', {
                    succeeded: String(succeededCount),
                    failed: String(failedCount),
                    message: failureMessage,
                }),
                { variant: 'warning' }
            );
        },
        [enqueueSnackbar, t]
    );

    const runAndNotify = useCallback(
        async (
            files: FileData[],
            operation: (file: FileData) => Promise<unknown>,
            messages: OutcomeMessages
        ): Promise<BulkOutcome> => {
            const outcome = await runBulkOperation(files, operation);
            notifyOutcome(outcome, messages);
            return outcome;
        },
        [notifyOutcome]
    );

    const moveFiles = useCallback(
        (files: FileData[], destination: FolderPickerResult) =>
            runAndNotify(
                files,
                (file) => moveFile(file.id, destination.folderId, destination.path),
                moveMessages
            ),
        [moveFile, runAndNotify]
    );

    const copyFiles = useCallback(
        (files: FileData[], destination: FolderPickerResult) =>
            runAndNotify(
                files,
                (file) => copyFile(file.id, destination.folderId, destination.path),
                copyMessages
            ),
        [copyFile, runAndNotify]
    );

    const deleteFiles = useCallback(
        (files: FileData[]) => runAndNotify(files, (file) => deleteFile(file.id), deleteMessages),
        [deleteFile, runAndNotify]
    );

    const renameSingleFile = useCallback(
        (file: FileData, newName: string) =>
            runAndNotify([file], (target) => renameFile(target.id, newName), renameMessages),
        [renameFile, runAndNotify]
    );

    const toggleFavorites = useCallback(
        (files: FileData[]) => {
            const shouldUnfavorite = files.every((file) => file.starred);
            const filesToToggle = shouldUnfavorite ? files : files.filter((file) => !file.starred);
            return runAndNotify(filesToToggle, (file) => toggleStarred(file.id), favoriteMessages);
        },
        [toggleStarred, runAndNotify]
    );

    const downloadFiles = useCallback((files: FileData[]) => {
        const [firstFile] = files;
        if (!firstFile) return;
        if (files.length === 1) {
            triggerBrowserDownload(getFileDownloadUrl(firstFile.id), firstFile.name);
            return;
        }
        triggerBrowserDownload(getFilesZipDownloadUrl(files.map((file) => file.id)));
    }, []);

    const copyPaths = useCallback(
        async (files: FileData[]) => {
            const clipboard = typeof navigator === 'undefined' ? undefined : navigator.clipboard;
            if (!clipboard?.writeText) {
                enqueueSnackbar(t('ERROR_COPY_PATH_FAILED'), { variant: 'error' });
                return;
            }
            try {
                await clipboard.writeText(files.map((file) => file.path).join('\n'));
                enqueueSnackbar(t('FILES_COPY_PATH_SUCCESS'), { variant: 'success' });
            } catch {
                enqueueSnackbar(t('ERROR_COPY_PATH_FAILED'), { variant: 'error' });
            }
        },
        [enqueueSnackbar, t]
    );

    return {
        moveFiles,
        copyFiles,
        deleteFiles,
        renameSingleFile,
        toggleFavorites,
        downloadFiles,
        copyPaths,
    };
};

export default useFileOperations;
