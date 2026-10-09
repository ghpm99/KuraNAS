import { IconButton } from '@mui/material';
import { Star } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getMusicTitle } from '@/utils/music';
import type { IMusicData } from '@/features/music/providers/musicProvider/musicProvider';
import useTrackStar from './useTrackStar';

interface TrackStarButtonProps {
    track: IMusicData;
    iconSize?: number;
    isRevealedOnRowHover?: boolean;
}

const TrackStarButton = ({
    track,
    iconSize = 16,
    isRevealedOnRowHover = false,
}: TrackStarButtonProps) => {
    const { t } = useI18n();
    const { isStarred, toggleStar } = useTrackStar(track);
    const trackTitle = getMusicTitle(track);

    return (
        <IconButton
            size="small"
            aria-label={t(isStarred ? 'MUSIC_TRACK_UNSTAR' : 'MUSIC_TRACK_STAR', {
                name: trackTitle,
            })}
            aria-pressed={isStarred}
            onClick={(event) => {
                event.stopPropagation();
                void toggleStar();
            }}
            sx={{
                color: isStarred ? 'warning.main' : 'text.secondary',
                opacity: isRevealedOnRowHover && !isStarred ? 0 : 1,
                '.MuiListItemButton-root:hover &, &:focus-visible': { opacity: 1 },
                '&:hover': { color: 'warning.main' },
            }}
        >
            <Star size={iconSize} fill={isStarred ? 'currentColor' : 'none'} />
        </IconButton>
    );
};

export default TrackStarButton;
