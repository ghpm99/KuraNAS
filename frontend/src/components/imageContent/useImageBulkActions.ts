import { useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import type { FolderPickerResult } from '@/components/folderPicker/folderPicker';
import {
    deleteFile,
    getFileDownloadUrl,
    getFilesZipDownloadUrl,
    moveFile,
    toggleStarredFile,
} from '@/service/files';
import { triggerBrowserDownload } from '@/service/browserDownload';
import { runBulkOperation } from '@/shared/bulk/runBulkOperation';
import {
    useBulkOutcomeNotifier,
    type OutcomeMessages,
} from '@/shared/bulk/useBulkOutcomeNotifier';
import { allFileQueryKeys } from '@/shared/queryKeys/fileQueryKeys';
import type { ImageLibraryItem } from '@/types/imageLibrary';

const imageLibraryQueryKeys = ['library', 'count', 'timeline', 'folders'];

const moveMessages: OutcomeMessages = {
    singleSuccessKey: 'ACTION_MOVE_SUCCESS',
    multipleSuccessKey: 'FILES_BULK_MOVE_SUCCESS',
    failureKey: 'ERROR_MOVE_FAILED',
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

export const shouldUnfavorite = (images: ImageLibraryItem[]) =>
    images.length > 0 && images.every((image) => image.starred);

export const useImageBulkActions = (onSucceeded: (images: ImageLibraryItem[]) => void) => {
    const queryClient = useQueryClient();
    const notifyOutcome = useBulkOutcomeNotifier();

    const invalidateLibraryAndFiles = useCallback(
        () =>
            Promise.all([
                ...imageLibraryQueryKeys.map((segment) =>
                    queryClient.invalidateQueries({ queryKey: ['images', segment] })
                ),
                ...allFileQueryKeys.map((queryKey) =>
                    queryClient.invalidateQueries({ queryKey: [queryKey] })
                ),
            ]),
        [queryClient]
    );

    const runAndNotify = useCallback(
        async (
            images: ImageLibraryItem[],
            operation: (image: ImageLibraryItem) => Promise<unknown>,
            messages: OutcomeMessages
        ) => {
            const outcome = await runBulkOperation(images, operation);
            notifyOutcome(
                {
                    succeededCount: outcome.succeededItems.length,
                    failedCount: outcome.failedItems.length,
                    firstFailureMessage: outcome.firstFailureMessage,
                },
                messages
            );
            onSucceeded(outcome.succeededItems);
            await invalidateLibraryAndFiles();
            return outcome;
        },
        [notifyOutcome, onSucceeded, invalidateLibraryAndFiles]
    );

    const moveImages = useCallback(
        (images: ImageLibraryItem[], destination: FolderPickerResult) =>
            runAndNotify(
                images,
                (image) => moveFile(image.file_id, destination.folderId, destination.path),
                moveMessages
            ),
        [runAndNotify]
    );

    const deleteImages = useCallback(
        (images: ImageLibraryItem[], isPermanent: boolean) =>
            runAndNotify(images, (image) => deleteFile(image.file_id, isPermanent), deleteMessages),
        [runAndNotify]
    );

    const toggleFavorites = useCallback(
        (images: ImageLibraryItem[]) => {
            const imagesToToggle = shouldUnfavorite(images)
                ? images
                : images.filter((image) => !image.starred);
            return runAndNotify(
                imagesToToggle,
                (image) => toggleStarredFile(image.file_id),
                favoriteMessages
            );
        },
        [runAndNotify]
    );

    const downloadImages = useCallback((images: ImageLibraryItem[]) => {
        const [firstImage] = images;
        if (!firstImage) {
            return;
        }
        if (images.length === 1) {
            triggerBrowserDownload(getFileDownloadUrl(firstImage.file_id), firstImage.name);
            return;
        }
        triggerBrowserDownload(getFilesZipDownloadUrl(images.map((image) => image.file_id)));
    }, []);

    return { moveImages, deleteImages, toggleFavorites, downloadImages };
};
