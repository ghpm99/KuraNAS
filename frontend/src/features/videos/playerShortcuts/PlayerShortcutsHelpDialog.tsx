import {
    Dialog,
    DialogContent,
    DialogTitle,
    Table,
    TableBody,
    TableCell,
    TableRow,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import { usePlayerPortalContainer } from '@/features/videos/videoPlayer/playerPortalContainer';
import { playerShortcutHelpEntries } from './playerShortcutHelpEntries';

interface PlayerShortcutsHelpDialogProps {
    isOpen?: boolean;
    onClose?: () => void;
}

export default function PlayerShortcutsHelpDialog({
    isOpen = false,
    onClose,
}: PlayerShortcutsHelpDialogProps) {
    const { t } = useI18n();
    const portalContainer = usePlayerPortalContainer();

    const renderKeyLabel = (key: string) => (key.startsWith('VIDEO_SHORTCUT_KEY_') ? t(key) : key);

    return (
        <Dialog
            open={isOpen}
            onClose={onClose}
            container={portalContainer}
            aria-labelledby="player-shortcuts-help-title"
        >
            <DialogTitle id="player-shortcuts-help-title">
                {t('VIDEO_SHORTCUT_HELP_TITLE')}
            </DialogTitle>
            <DialogContent>
                <Table size="small">
                    <TableBody>
                        {playerShortcutHelpEntries.map((entry) => (
                            <TableRow key={entry.descriptionKey}>
                                <TableCell>
                                    {entry.keys.map((key) => (
                                        <kbd key={key} style={{ marginRight: 6 }}>
                                            {renderKeyLabel(key)}
                                        </kbd>
                                    ))}
                                </TableCell>
                                <TableCell>{t(entry.descriptionKey)}</TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </DialogContent>
        </Dialog>
    );
}
