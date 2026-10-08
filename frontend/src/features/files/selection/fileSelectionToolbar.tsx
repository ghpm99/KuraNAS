import { Box, Button, IconButton, Tooltip, Typography } from '@mui/material';
import { CheckCheck, Copy, Download, MoveRight, Pencil, Star, Trash2, X } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import useFile from '@/features/files/providers/fileProvider/fileContext';
import useFileActionFlow from '@/features/files/fileActions/useFileActionFlow';
import useFileOperations from '@/features/files/fileActions/useFileOperations';
import { useFileSelectionContext } from './fileSelectionContext';
import { resolveListedFiles } from './listedFiles';

const FileSelectionToolbar = () => {
    const { t } = useI18n();
    const { selectedItem, files } = useFile();
    const { selectedFiles, selectedCount, selectAll, clear, deselect } = useFileSelectionContext();
    const { startAction, dialogs } = useFileActionFlow();
    const { downloadFiles, toggleFavorites } = useFileOperations();

    const allSelectedAreFavorites = selectedFiles.every((file) => file.starred);
    const canRename = selectedCount === 1;

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
                <IconButton
                    size="small"
                    aria-label={t('FILES_CLEAR_SELECTION')}
                    onClick={clear}
                >
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
                onClick={() => selectAll(resolveListedFiles(selectedItem, files))}
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
                startIcon={<Copy size={16} />}
                onClick={() => startAction('copy', selectedFiles)}
            >
                {t('COPY')}
            </Button>
            <Button
                size="small"
                variant="outlined"
                startIcon={<Star size={16} fill={allSelectedAreFavorites ? 'currentColor' : 'none'} />}
                onClick={favoriteSelected}
            >
                {allSelectedAreFavorites ? t('FILES_UNFAVORITE') : t('FILES_FAVORITE')}
            </Button>
            {canRename ? (
                <Button
                    size="small"
                    variant="outlined"
                    startIcon={<Pencil size={16} />}
                    onClick={() => startAction('rename', selectedFiles)}
                >
                    {t('RENAME')}
                </Button>
            ) : null}
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

export default FileSelectionToolbar;
