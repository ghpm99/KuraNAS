import { useState } from 'react';
import { Box, Button, IconButton, Tooltip, Typography } from '@mui/material';
import { CheckCheck, Download, MoveRight, Star, Trash2, X } from 'lucide-react';
import DeleteItemsDialog from '@/components/deleteItemsDialog/deleteItemsDialog';
import FolderPicker from '@/components/folderPicker/folderPicker';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { ImageLibraryItem } from '@/types/imageLibrary';
import { shouldUnfavorite, useImageBulkActions } from '../useImageBulkActions';
import type { ImageSelection } from '../useImageSelection';
import styles from '../ImageContent.module.css';

type ImageSelectionToolbarProps = {
    selection: ImageSelection;
    loadedImages: ImageLibraryItem[];
};

type PendingDialog = 'move' | 'delete' | null;

export default function ImageSelectionToolbar({
    selection,
    loadedImages,
}: ImageSelectionToolbarProps) {
    const { t } = useI18n();
    const [pendingDialog, setPendingDialog] = useState<PendingDialog>(null);
    const { selectedItems, selectedCount, deselect, selectAll, clear } = selection;
    const { moveImages, deleteImages, toggleFavorites, downloadImages } = useImageBulkActions(deselect);

    const closeDialog = () => setPendingDialog(null);
    const isUnfavoriteAction = shouldUnfavorite(selectedItems);

    return (
        <Box
            role="toolbar"
            aria-label={t('IMAGES_SELECTION_TOOLBAR')}
            className={styles.selectionToolbar}
        >
            <Tooltip title={t('FILES_CLEAR_SELECTION')}>
                <IconButton size="small" aria-label={t('FILES_CLEAR_SELECTION')} onClick={clear}>
                    <X size={16} />
                </IconButton>
            </Tooltip>
            <Typography variant="body2" fontWeight={600} sx={{ mr: 1 }}>
                {t('FILES_SELECTION_COUNT', { count: String(selectedCount) })}
            </Typography>
            <Button
                size="small"
                variant="outlined"
                startIcon={<CheckCheck size={16} />}
                onClick={() => selectAll(loadedImages)}
            >
                {t('IMAGES_SELECT_ALL_LOADED')}
            </Button>
            <Button
                size="small"
                variant="outlined"
                startIcon={<Download size={16} />}
                onClick={() => downloadImages(selectedItems)}
            >
                {t('DOWNLOAD')}
            </Button>
            <Button
                size="small"
                variant="outlined"
                startIcon={<Star size={16} fill={isUnfavoriteAction ? 'currentColor' : 'none'} />}
                onClick={() => toggleFavorites(selectedItems)}
            >
                {isUnfavoriteAction ? t('FILES_UNFAVORITE') : t('FILES_FAVORITE')}
            </Button>
            <Button
                size="small"
                variant="outlined"
                startIcon={<MoveRight size={16} />}
                onClick={() => setPendingDialog('move')}
            >
                {t('MOVE')}
            </Button>
            <Button
                size="small"
                color="error"
                variant="outlined"
                startIcon={<Trash2 size={16} />}
                onClick={() => setPendingDialog('delete')}
            >
                {t('DELETE')}
            </Button>
            <FolderPicker
                open={pendingDialog === 'move'}
                mode="move"
                onClose={closeDialog}
                onSelect={async (destination) => {
                    closeDialog();
                    await moveImages(selectedItems, destination);
                }}
            />
            <DeleteItemsDialog
                items={selectedItems}
                isOpen={pendingDialog === 'delete'}
                onClose={closeDialog}
                onConfirm={async (images, isPermanent) => {
                    closeDialog();
                    await deleteImages(images, isPermanent);
                }}
            />
        </Box>
    );
}
