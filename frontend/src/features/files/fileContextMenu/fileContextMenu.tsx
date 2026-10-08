import { Divider, ListItemIcon, ListItemText, Menu, MenuItem } from '@mui/material';
import { Copy, Download, FolderOpen, Link, MoveRight, Pencil, Star, Trash2 } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import useFileActionFlow, {
    type DialogFileAction,
} from '@/features/files/fileActions/useFileActionFlow';
import useFileOperations from '@/features/files/fileActions/useFileOperations';
import { useFileSelectionContext } from '@/features/files/selection/fileSelectionContext';

export type FileContextMenuAnchor = { top: number; left: number };

type FileContextMenuProps = {
    anchorPosition: FileContextMenuAnchor | null;
    targetFiles: FileData[];
    onClose: () => void;
    onOpenFile: (file: FileData) => void;
};

const FileContextMenu = ({
    anchorPosition,
    targetFiles,
    onClose,
    onOpenFile,
}: FileContextMenuProps) => {
    const { t } = useI18n();
    const { startAction, dialogs } = useFileActionFlow();
    const { downloadFiles, toggleFavorites, copyPaths } = useFileOperations();
    const { deselect } = useFileSelectionContext();

    const isSingleTarget = targetFiles.length === 1;
    const allTargetsAreFavorites = targetFiles.every((file) => file.starred);

    const runAndClose = (action: () => void) => () => {
        onClose();
        action();
    };

    const startDialogAction = (action: DialogFileAction) =>
        runAndClose(() => startAction(action, targetFiles));

    const favoriteTargets = async () => {
        const outcome = await toggleFavorites(targetFiles);
        deselect(outcome.succeededFiles);
    };

    const [firstTarget] = targetFiles;

    return (
        <>
            <Menu
                open={anchorPosition !== null && targetFiles.length > 0}
                onClose={onClose}
                anchorReference="anchorPosition"
                anchorPosition={anchorPosition ?? undefined}
            >
                <MenuItem
                    disabled={!isSingleTarget}
                    onClick={runAndClose(() => firstTarget && onOpenFile(firstTarget))}
                >
                    <ListItemIcon>
                        <FolderOpen size={16} />
                    </ListItemIcon>
                    <ListItemText>{t('FILES_OPEN')}</ListItemText>
                </MenuItem>
                <MenuItem disabled={!isSingleTarget} onClick={startDialogAction('rename')}>
                    <ListItemIcon>
                        <Pencil size={16} />
                    </ListItemIcon>
                    <ListItemText>{t('RENAME')}</ListItemText>
                </MenuItem>
                <MenuItem onClick={startDialogAction('move')}>
                    <ListItemIcon>
                        <MoveRight size={16} />
                    </ListItemIcon>
                    <ListItemText>{t('MOVE')}</ListItemText>
                </MenuItem>
                <MenuItem onClick={startDialogAction('copy')}>
                    <ListItemIcon>
                        <Copy size={16} />
                    </ListItemIcon>
                    <ListItemText>{t('COPY')}</ListItemText>
                </MenuItem>
                <MenuItem onClick={runAndClose(() => downloadFiles(targetFiles))}>
                    <ListItemIcon>
                        <Download size={16} />
                    </ListItemIcon>
                    <ListItemText>{t('DOWNLOAD')}</ListItemText>
                </MenuItem>
                <MenuItem onClick={runAndClose(favoriteTargets)}>
                    <ListItemIcon>
                        <Star size={16} fill={allTargetsAreFavorites ? 'currentColor' : 'none'} />
                    </ListItemIcon>
                    <ListItemText>
                        {allTargetsAreFavorites ? t('FILES_UNFAVORITE') : t('FILES_FAVORITE')}
                    </ListItemText>
                </MenuItem>
                <MenuItem onClick={runAndClose(() => copyPaths(targetFiles))}>
                    <ListItemIcon>
                        <Link size={16} />
                    </ListItemIcon>
                    <ListItemText>{t('FILES_COPY_PATH')}</ListItemText>
                </MenuItem>
                <Divider />
                <MenuItem onClick={startDialogAction('delete')}>
                    <ListItemIcon>
                        <Trash2 size={16} />
                    </ListItemIcon>
                    <ListItemText>{t('DELETE')}</ListItemText>
                </MenuItem>
            </Menu>
            {dialogs}
        </>
    );
};

export default FileContextMenu;
