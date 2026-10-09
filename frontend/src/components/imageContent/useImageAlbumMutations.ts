import { useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import {
    addImageAlbumItems,
    createImageAlbum,
    deleteImageAlbum,
    removeImageAlbumItems,
    updateImageAlbum,
} from '@/service/imageAlbum';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import type { ImageAlbum } from '@/types/imageAlbum';

const albumQueryPrefixes = ['albums', 'library', 'count', 'timeline'];

export const useImageAlbumMutations = () => {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const queryClient = useQueryClient();

    const refreshAlbumQueries = useCallback(
        () =>
            Promise.all(
                albumQueryPrefixes.map((prefix) =>
                    queryClient.invalidateQueries({ queryKey: ['images', prefix] })
                )
            ),
        [queryClient]
    );

    const runAlbumAction = useCallback(
        async <TResult>(
            action: () => Promise<TResult>,
            buildSuccessMessage: (result: TResult) => string
        ): Promise<TResult | undefined> => {
            try {
                const result = await action();
                enqueueSnackbar(buildSuccessMessage(result), { variant: 'success' });
                await refreshAlbumQueries();
                return result;
            } catch (error) {
                enqueueSnackbar(
                    extractBackendErrorMessage(error) ?? t('IMAGES_ALBUM_ERROR_GENERIC'),
                    {
                        variant: 'error',
                    }
                );
                return undefined;
            }
        },
        [enqueueSnackbar, refreshAlbumQueries, t]
    );

    const createAlbum = useCallback(
        (name: string) =>
            runAlbumAction(
                () => createImageAlbum(name),
                (album) => t('IMAGES_ALBUM_CREATED', { name: album.name })
            ),
        [runAlbumAction, t]
    );

    const renameAlbum = useCallback(
        (albumId: number, name: string) =>
            runAlbumAction(
                () => updateImageAlbum(albumId, { name }),
                (album) => t('IMAGES_ALBUM_RENAMED', { name: album.name })
            ),
        [runAlbumAction, t]
    );

    const setAlbumCover = useCallback(
        (albumId: number, coverFileId: number) =>
            runAlbumAction(
                () => updateImageAlbum(albumId, { cover_file_id: coverFileId }),
                () => t('IMAGES_ALBUM_COVER_SET')
            ),
        [runAlbumAction, t]
    );

    const deleteAlbum = useCallback(
        (album: ImageAlbum) =>
            runAlbumAction(
                () => deleteImageAlbum(album.id),
                () => t('IMAGES_ALBUM_DELETED', { name: album.name })
            ),
        [runAlbumAction, t]
    );

    const addPhotos = useCallback(
        (album: ImageAlbum, fileIds: number[]) =>
            runAlbumAction(
                () => addImageAlbumItems(album.id, fileIds),
                (change) => {
                    const skippedCount = change.requested - change.changed;
                    return skippedCount === 0
                        ? t('IMAGES_ALBUM_ITEMS_ADDED', {
                              count: String(change.changed),
                              name: album.name,
                          })
                        : t('IMAGES_ALBUM_ITEMS_ADDED_PARTIAL', {
                              count: String(change.changed),
                              skipped: String(skippedCount),
                              name: album.name,
                          });
                }
            ),
        [runAlbumAction, t]
    );

    const removePhotos = useCallback(
        (albumId: number, fileIds: number[]) =>
            runAlbumAction(
                () => removeImageAlbumItems(albumId, fileIds),
                (change) => t('IMAGES_ALBUM_ITEMS_REMOVED', { count: String(change.changed) })
            ),
        [runAlbumAction, t]
    );

    return { createAlbum, renameAlbum, setAlbumCover, deleteAlbum, addPhotos, removePhotos };
};
