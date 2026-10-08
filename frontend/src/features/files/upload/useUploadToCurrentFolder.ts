import { useCallback } from 'react';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import useFile, { type FileData } from '@/features/files/providers/fileProvider/fileContext';
import { getFileByPath } from '@/service/files';
import { FileType } from '@/utils';
import useUploadQueueContext from './uploadQueueContext';
import type { UploadEntry } from './uploadQueueTypes';

const rootFolderPaths = ['', '/'];

export const resolveUploadFolderId = async (
    currentItem: FileData | null
): Promise<number | undefined> => {
    if (!currentItem) return undefined;
    if (currentItem.type === FileType.Directory) return currentItem.id;
    if (rootFolderPaths.includes(currentItem.parent_path ?? '')) return undefined;
    const parentFolder = await getFileByPath(currentItem.parent_path);
    if (!parentFolder) throw new Error('upload parent folder not found');
    return parentFolder.id;
};

export const useUploadToCurrentFolder = () => {
    const { selectedItem } = useFile();
    const { enqueue } = useUploadQueueContext();
    const { enqueueSnackbar } = useSnackbar();
    const { t } = useI18n();

    const uploadEntries = useCallback(
        async (entries: UploadEntry[]) => {
            if (entries.length === 0) return;
            try {
                const targetFolderId = await resolveUploadFolderId(selectedItem);
                enqueue(entries, targetFolderId);
            } catch {
                enqueueSnackbar(t('ERROR_UPLOAD_FAILED'), { variant: 'error' });
            }
        },
        [selectedItem, enqueue, enqueueSnackbar, t]
    );

    return { uploadEntries };
};

export default useUploadToCurrentFolder;
