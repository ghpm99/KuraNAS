import { Box, Button, IconButton, Tooltip, Typography } from '@mui/material';
import { CheckCheck, Download, MoveRight, Star, Trash2, X } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import useFileActionFlow from '@/features/files/fileActions/useFileActionFlow';
import useFileOperations from '@/features/files/fileActions/useFileOperations';
import { useFileSelectionContext } from '@/features/files/selection/fileSelectionContext';

type DocumentsBatchBarProps = {
    selectableFiles: FileData[];
};

const DocumentsBatchBar = ({ selectableFiles }: DocumentsBatchBarProps) => {
    const { t } = useI18n();
    const { selectedFiles, selectedCount, selectAll, clear, deselect } = useFileSelectionContext();
    const { startAction, dialogs } = useFileActionFlow();
    const { downloadFiles, toggleFavorites } = useFileOperations();

    const areAllSelectedFavorites = selectedFiles.every((file) => file.starred);

    const favoriteSelected = async () => {
        const outcome = await toggleFavorites(selectedFiles);
        deselect(outcome.succeededFiles);
    };

    return (
        <Box
            role="toolbar"
            aria-label={t('FILES_SELECTION_TOOLBAR')}
            sx={{ display: 'flex', flexWrap: 'wrap', alignItems: 'center', gap: 1, minWidth: 0 }}
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
                onClick={() => selectAll(selectableFiles)}
            >
                {t('FILES_SELECT_ALL')}
            </Button>
            <Button
                size="small"
                variant="outlined"
                startIcon={<Download size={16} />}
                onClick={() => downloadFiles(selectedFiles)}
            >
                {t('DOWNLOAD')}
            </Button>
            <Button
                size="small"
                variant="outlined"
                startIcon={<MoveRight size={16} />}
                onClick={() => startAction('move', selectedFiles)}
            >
                {t('MOVE')}
            </Button>
            <Button
                size="small"
                variant="outlined"
                startIcon={
                    <Star size={16} fill={areAllSelectedFavorites ? 'currentColor' : 'none'} />
                }
                onClick={favoriteSelected}
            >
                {areAllSelectedFavorites ? t('FILES_UNFAVORITE') : t('FILES_FAVORITE')}
            </Button>
            <Button
                size="small"
                color="error"
                variant="outlined"
                startIcon={<Trash2 size={16} />}
                onClick={() => startAction('delete', selectedFiles)}
            >
                {t('DELETE')}
            </Button>
            {dialogs}
        </Box>
    );
};

export default DocumentsBatchBar;
