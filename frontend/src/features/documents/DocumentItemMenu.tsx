import { Divider, ListItemIcon, ListItemText, Menu, MenuItem } from '@mui/material';
import { Download, Info, MoveRight, Pencil, Star, Trash2 } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import useFileActionFlow, {
    type DialogFileAction,
} from '@/features/files/fileActions/useFileActionFlow';
import useFileOperations from '@/features/files/fileActions/useFileOperations';
import useFileDetails from '@/features/files/fileDetails/useFileDetails';
import { useFileSelectionContext } from '@/features/files/selection/fileSelectionContext';

export type DocumentItemMenuAnchor = { top: number; left: number };

type DocumentItemMenuProps = {
    anchorPosition: DocumentItemMenuAnchor | null;
    targetFiles: FileData[];
    onClose: () => void;
};

const DocumentItemMenu = ({ anchorPosition, targetFiles, onClose }: DocumentItemMenuProps) => {
    const { t } = useI18n();
    const { startAction, dialogs } = useFileActionFlow();
    const { downloadFiles, toggleFavorites } = useFileOperations();
    const { deselect } = useFileSelectionContext();
    const { isAvailable: isDetailsAvailable, openDetails } = useFileDetails();

    const [firstTarget] = targetFiles;
    const isSingleTarget = targetFiles.length === 1;
    const areAllTargetsFavorites = targetFiles.every((file) => file.starred);

    const closeThenRun = (action: () => void) => () => {
        onClose();
        action();
    };

    const startDialogAction = (action: DialogFileAction) =>
        closeThenRun(() => startAction(action, targetFiles));

    const favoriteTargets = async () => {
        const outcome = await toggleFavorites(targetFiles);
        deselect(outcome.succeededFiles);
    };

    return (
        <>
            <Menu
                open={anchorPosition !== null && targetFiles.length > 0}
                onClose={onClose}
                anchorReference="anchorPosition"
                anchorPosition={anchorPosition ?? undefined}
            >
                <MenuItem onClick={closeThenRun(() => downloadFiles(targetFiles))}>
                    <ListItemIcon>
                        <Download size={16} />
                    </ListItemIcon>
                    <ListItemText>{t('DOWNLOAD')}</ListItemText>
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
                <MenuItem onClick={closeThenRun(favoriteTargets)}>
                    <ListItemIcon>
                        <Star size={16} fill={areAllTargetsFavorites ? 'currentColor' : 'none'} />
                    </ListItemIcon>
                    <ListItemText>
                        {areAllTargetsFavorites ? t('FILES_UNFAVORITE') : t('FILES_FAVORITE')}
                    </ListItemText>
                </MenuItem>
                {isDetailsAvailable ? (
                    <MenuItem
                        disabled={!isSingleTarget}
                        onClick={closeThenRun(() => firstTarget && openDetails(firstTarget))}
                    >
                        <ListItemIcon>
                            <Info size={16} />
                        </ListItemIcon>
                        <ListItemText>{t('FILES_DETAILS')}</ListItemText>
                    </MenuItem>
                ) : null}
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

export default DocumentItemMenu;
