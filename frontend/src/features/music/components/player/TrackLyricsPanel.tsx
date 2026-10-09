import { Box, CircularProgress, Typography } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { IMusicData } from '@/features/music/providers/musicProvider/musicProvider';
import { useTrackLyrics } from './useTrackLyrics';
import styles from './TrackLyricsPanel.module.css';

interface TrackLyricsPanelProps {
    track?: IMusicData;
}

const TrackLyricsPanel = ({ track }: TrackLyricsPanelProps) => {
    const { t } = useI18n();
    const { lyrics, isLoading } = useTrackLyrics(track, true);

    if (isLoading) {
        return (
            <Box className={styles.panel} role="status" aria-label={t('PLAYER_LYRICS_LOADING')}>
                <CircularProgress size={24} />
            </Box>
        );
    }

    if (lyrics === '') {
        return (
            <Box className={styles.panel}>
                <Typography variant="body2" color="text.secondary">
                    {t('PLAYER_LYRICS_UNAVAILABLE')}
                </Typography>
            </Box>
        );
    }

    return (
        <Box className={styles.panel} data-testid="track-lyrics">
            <Typography variant="body1" className={styles.lyricsText}>
                {lyrics}
            </Typography>
        </Box>
    );
};

export default TrackLyricsPanel;
