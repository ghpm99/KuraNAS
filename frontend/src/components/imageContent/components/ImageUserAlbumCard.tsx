import { useState } from 'react';
import { IconButton, Menu, MenuItem } from '@mui/material';
import { MoreVertical } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { ImageAlbum } from '@/types/imageAlbum';
import {
    GRID_THUMBNAIL_SIZE,
    gridThumbnailSizes,
    thumbnailSrcSet,
    thumbnailUrl,
} from '../imageThumbnailSources';
import collectionStyles from './ImageCollectionsPanel.module.css';
import styles from './ImageUserAlbums.module.css';

type ImageUserAlbumCardProps = {
    album: ImageAlbum;
    onOpen: (albumId: number) => void;
    onRename: (album: ImageAlbum) => void;
    onDelete: (album: ImageAlbum) => void;
};

export default function ImageUserAlbumCard({
    album,
    onOpen,
    onRename,
    onDelete,
}: ImageUserAlbumCardProps) {
    const { t } = useI18n();
    const [menuAnchor, setMenuAnchor] = useState<HTMLElement | null>(null);
    const coverFileId = album.cover_file_id ?? undefined;

    const chooseMenuAction = (action: (album: ImageAlbum) => void) => {
        setMenuAnchor(null);
        action(album);
    };

    return (
        <div className={styles.cardSlot}>
            <button
                type="button"
                className={collectionStyles.card}
                onClick={() => onOpen(album.id)}
                aria-label={t('IMAGES_COLLECTION_OPEN', { name: album.name })}
            >
                <div className={collectionStyles.cover}>
                    {coverFileId ? (
                        <img
                            src={thumbnailUrl(coverFileId, GRID_THUMBNAIL_SIZE)}
                            srcSet={thumbnailSrcSet(coverFileId, GRID_THUMBNAIL_SIZE)}
                            sizes={gridThumbnailSizes}
                            alt={album.name}
                            loading="lazy"
                        />
                    ) : (
                        <div className={collectionStyles.placeholder}>
                            {album.name.slice(0, 1).toUpperCase()}
                        </div>
                    )}
                </div>
                <div className={collectionStyles.body}>
                    <div className={collectionStyles.meta}>
                        <h3>{album.name}</h3>
                    </div>
                    <span className={collectionStyles.count}>
                        {t('IMAGES_PHOTOS_COUNT', { count: String(album.item_count) })}
                    </span>
                </div>
            </button>
            <IconButton
                size="small"
                className={styles.menuButton}
                aria-label={t('IMAGES_ALBUM_MENU_ARIA', { name: album.name })}
                onClick={(event) => setMenuAnchor(event.currentTarget)}
            >
                <MoreVertical size={16} />
            </IconButton>
            <Menu
                anchorEl={menuAnchor}
                open={Boolean(menuAnchor)}
                onClose={() => setMenuAnchor(null)}
            >
                <MenuItem onClick={() => chooseMenuAction(onRename)}>{t('RENAME')}</MenuItem>
                <MenuItem onClick={() => chooseMenuAction(onDelete)}>{t('DELETE')}</MenuItem>
            </Menu>
        </div>
    );
}
