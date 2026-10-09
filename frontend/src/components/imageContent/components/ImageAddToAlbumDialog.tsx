import { useState } from 'react';
import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    List,
    ListItemButton,
    ListItemText,
    TextField,
    Typography,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { ImageAlbum } from '@/types/imageAlbum';
import { useImageAlbumMutations } from '../useImageAlbumMutations';
import { useUserImageAlbums } from '../useUserImageAlbums';
import styles from './ImageUserAlbums.module.css';

type ImageAddToAlbumDialogProps = {
    isOpen: boolean;
    onClose: () => void;
    onPickAlbum: (album: ImageAlbum) => void;
};

const ImageAddToAlbumDialogBody = ({
    onClose,
    onPickAlbum,
}: Omit<ImageAddToAlbumDialogProps, 'isOpen'>) => {
    const { t } = useI18n();
    const [searchText, setSearchText] = useState('');
    const [newAlbumName, setNewAlbumName] = useState('');
    const { albums } = useUserImageAlbums(true);
    const { createAlbum } = useImageAlbumMutations();
    const trimmedNewAlbumName = newAlbumName.trim();
    const normalizedSearch = searchText.trim().toLowerCase();
    const matchingAlbums = albums.filter((album) =>
        album.name.toLowerCase().includes(normalizedSearch)
    );

    const createAndPickAlbum = async () => {
        const createdAlbum = await createAlbum(trimmedNewAlbumName);
        if (createdAlbum) {
            setNewAlbumName('');
            onPickAlbum(createdAlbum);
        }
    };

    return (
        <>
            <DialogTitle>{t('IMAGES_ALBUM_ADD_DIALOG_TITLE')}</DialogTitle>
            <DialogContent>
                <TextField
                    fullWidth
                    size="small"
                    margin="dense"
                    placeholder={t('IMAGES_ALBUM_SEARCH_PLACEHOLDER')}
                    value={searchText}
                    onChange={(event) => setSearchText(event.target.value)}
                    slotProps={{
                        htmlInput: { 'aria-label': t('IMAGES_ALBUM_SEARCH_PLACEHOLDER') },
                    }}
                />
                <List dense className={styles.albumList}>
                    {matchingAlbums.map((album) => (
                        <ListItemButton key={album.id} onClick={() => onPickAlbum(album)}>
                            <ListItemText
                                primary={album.name}
                                secondary={t('IMAGES_PHOTOS_COUNT', {
                                    count: String(album.item_count),
                                })}
                            />
                        </ListItemButton>
                    ))}
                </List>
                {matchingAlbums.length === 0 && (
                    <Typography variant="body2" color="text.secondary">
                        {t('IMAGES_ALBUM_NO_RESULTS')}
                    </Typography>
                )}
                <TextField
                    fullWidth
                    size="small"
                    margin="normal"
                    label={t('IMAGES_ALBUM_NAME_LABEL')}
                    value={newAlbumName}
                    onChange={(event) => setNewAlbumName(event.target.value)}
                />
                <Button
                    variant="outlined"
                    size="small"
                    disabled={!trimmedNewAlbumName}
                    onClick={createAndPickAlbum}
                >
                    {t('IMAGES_ALBUM_CREATE')}
                </Button>
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>{t('ACTION_CANCEL')}</Button>
            </DialogActions>
        </>
    );
};

export default function ImageAddToAlbumDialog({
    isOpen,
    ...bodyProps
}: ImageAddToAlbumDialogProps) {
    return (
        <Dialog open={isOpen} onClose={bodyProps.onClose} fullWidth maxWidth="xs">
            {isOpen && <ImageAddToAlbumDialogBody {...bodyProps} />}
        </Dialog>
    );
}
