import { Box, IconButton } from '@mui/material';
import { ArrowDown, ArrowUp } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';

type PlaylistTrackMoveButtonsProps = {
    canMoveUp: boolean;
    canMoveDown: boolean;
    onMoveUp: () => void;
    onMoveDown: () => void;
};

export default function PlaylistTrackMoveButtons({
    canMoveUp,
    canMoveDown,
    onMoveUp,
    onMoveDown,
}: PlaylistTrackMoveButtonsProps) {
    const { t } = useI18n();

    return (
        <Box
            sx={{ display: 'flex', flexShrink: 0 }}
            onClick={(event) => event.stopPropagation()}
            onKeyDown={(event) => event.stopPropagation()}
        >
            <IconButton
                size="small"
                aria-label={t('MUSIC_PLAYLIST_MOVE_UP')}
                disabled={!canMoveUp}
                onClick={onMoveUp}
            >
                <ArrowUp size={14} />
            </IconButton>
            <IconButton
                size="small"
                aria-label={t('MUSIC_PLAYLIST_MOVE_DOWN')}
                disabled={!canMoveDown}
                onClick={onMoveDown}
            >
                <ArrowDown size={14} />
            </IconButton>
        </Box>
    );
}
