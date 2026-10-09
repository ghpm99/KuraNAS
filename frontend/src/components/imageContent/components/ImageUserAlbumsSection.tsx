import { useState } from 'react';
import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    Typography,
} from '@mui/material';
import { Plus } from 'lucide-react';
import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { ImageAlbum } from '@/types/imageAlbum';
import { useImageAlbumMutations } from '../useImageAlbumMutations';
import { useUserImageAlbums } from '../useUserImageAlbums';
import ImageAlbumNameDialog from './ImageAlbumNameDialog';
import ImageUserAlbumCard from './ImageUserAlbumCard';
import collectionStyles from './ImageCollectionsPanel.module.css';
import styles from './ImageUserAlbums.module.css';

type ImageUserAlbumsSectionProps = {
    onOpenAlbum: (albumId: number) => void;
};

type AlbumDialog =
    | { kind: 'create' }
    | { kind: 'rename'; album: ImageAlbum }
    | { kind: 'delete'; album: ImageAlbum }
    | null;

export default function ImageUserAlbumsSection({ onOpenAlbum }: ImageUserAlbumsSectionProps) {
    const { t } = useI18n();
    const [dialog, setDialog] = useState<AlbumDialog>(null);
    const { albums, status, hasNextPage, isFetchingNextPage, fetchNextPage } =
        useUserImageAlbums(true);
    const { createAlbum, renameAlbum, deleteAlbum } = useImageAlbumMutations();

    const closeDialog = () => setDialog(null);

    return (
        <section className={styles.section} aria-label={t('IMAGES_MY_ALBUMS_TITLE')}>
            <div className={styles.header}>
                <h3>{t('IMAGES_MY_ALBUMS_TITLE')}</h3>
                <Button
                    size="small"
                    variant="outlined"
                    startIcon={<Plus size={16} />}
                    onClick={() => setDialog({ kind: 'create' })}
                >
                    {t('IMAGES_ALBUM_CREATE')}
                </Button>
            </div>
            {status === 'error' && (
                <Typography variant="body2">{t('IMAGES_ALBUM_LOAD_ERROR')}</Typography>
            )}
            {status === 'success' && albums.length === 0 && (
                <div className={collectionStyles.empty}>
                    <h3>{t('IMAGES_USER_ALBUMS_EMPTY_TITLE')}</h3>
                    <p>{t('IMAGES_USER_ALBUMS_EMPTY_DESC')}</p>
                </div>
            )}
            {albums.length > 0 && (
                <div className={collectionStyles.grid}>
                    {albums.map((album) => (
                        <ImageUserAlbumCard
                            key={album.id}
                            album={album}
                            onOpen={onOpenAlbum}
                            onRename={(target) => setDialog({ kind: 'rename', album: target })}
                            onDelete={(target) => setDialog({ kind: 'delete', album: target })}
                        />
                    ))}
                </div>
            )}
            <LoadMoreSentinel
                hasNextPage={hasNextPage}
                isFetchingNextPage={isFetchingNextPage}
                fetchNextPage={() => fetchNextPage()}
            />
            <ImageAlbumNameDialog
                isOpen={dialog?.kind === 'create'}
                title={t('IMAGES_ALBUM_CREATE')}
                confirmLabel={t('IMAGES_ALBUM_CREATE_CONFIRM')}
                onClose={closeDialog}
                onSubmit={(name) => {
                    closeDialog();
                    createAlbum(name);
                }}
            />
            <ImageAlbumNameDialog
                isOpen={dialog?.kind === 'rename'}
                title={t('IMAGES_ALBUM_RENAME_TITLE')}
                confirmLabel={t('IMAGES_ALBUM_RENAME_CONFIRM')}
                initialName={dialog?.kind === 'rename' ? dialog.album.name : ''}
                onClose={closeDialog}
                onSubmit={(name) => {
                    if (dialog?.kind === 'rename') {
                        renameAlbum(dialog.album.id, name);
                    }
                    closeDialog();
                }}
            />
            <Dialog open={dialog?.kind === 'delete'} onClose={closeDialog}>
                <DialogTitle>{t('IMAGES_ALBUM_DELETE_TITLE')}</DialogTitle>
                <DialogContent>
                    <Typography variant="body2">
                        {dialog?.kind === 'delete'
                            ? t('IMAGES_ALBUM_DELETE_CONFIRM', { name: dialog.album.name })
                            : ''}
                    </Typography>
                </DialogContent>
                <DialogActions>
                    <Button onClick={closeDialog}>{t('ACTION_CANCEL')}</Button>
                    <Button
                        color="error"
                        variant="contained"
                        onClick={() => {
                            if (dialog?.kind === 'delete') {
                                deleteAlbum(dialog.album);
                            }
                            closeDialog();
                        }}
                    >
                        {t('DELETE')}
                    </Button>
                </DialogActions>
            </Dialog>
        </section>
    );
}
