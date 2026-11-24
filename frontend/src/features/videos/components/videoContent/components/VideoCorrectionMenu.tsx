import { IconButton, Menu, MenuItem } from '@mui/material';
import { MoreVertical } from 'lucide-react';
import { useState, type MouseEvent } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import VideoClassificationDialog from './VideoClassificationDialog';
import VideoEpisodeDialog from './VideoEpisodeDialog';

type VideoCorrectionMenuProps = {
    fileId: number;
    fileName: string;
    canCorrectEpisode?: boolean;
    season?: number | null;
    episode?: number | null;
};

type OpenDialog = 'classification' | 'episode' | null;

export default function VideoCorrectionMenu({
    fileId,
    fileName,
    canCorrectEpisode = false,
    season,
    episode,
}: VideoCorrectionMenuProps) {
    const { t } = useI18n();
    const [menuAnchor, setMenuAnchor] = useState<HTMLElement | null>(null);
    const [openDialog, setOpenDialog] = useState<OpenDialog>(null);

    const openMenu = (event: MouseEvent<HTMLElement>) => {
        event.stopPropagation();
        setMenuAnchor(event.currentTarget);
    };

    const chooseDialog = (dialog: OpenDialog) => {
        setMenuAnchor(null);
        setOpenDialog(dialog);
    };

    return (
        <>
            <IconButton
                size="small"
                aria-label={t('VIDEO_CORRECTION_MENU', { name: fileName })}
                onClick={openMenu}
            >
                <MoreVertical size={16} />
            </IconButton>
            <Menu
                anchorEl={menuAnchor}
                open={menuAnchor !== null}
                onClose={() => setMenuAnchor(null)}
            >
                <MenuItem onClick={() => chooseDialog('classification')}>
                    {t('VIDEO_CORRECT_CLASSIFICATION')}
                </MenuItem>
                {canCorrectEpisode && (
                    <MenuItem onClick={() => chooseDialog('episode')}>
                        {t('VIDEO_CORRECT_EPISODE')}
                    </MenuItem>
                )}
            </Menu>
            {openDialog === 'classification' && (
                <VideoClassificationDialog
                    isOpen
                    fileId={fileId}
                    fileName={fileName}
                    onClose={() => setOpenDialog(null)}
                />
            )}
            {openDialog === 'episode' && (
                <VideoEpisodeDialog
                    isOpen
                    fileId={fileId}
                    fileName={fileName}
                    initialSeason={season}
                    initialEpisode={episode}
                    onClose={() => setOpenDialog(null)}
                />
            )}
        </>
    );
}
