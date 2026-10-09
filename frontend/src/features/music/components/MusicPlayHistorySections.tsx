import { Button } from '@mui/material';
import { Music } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import CoverArt from '@/features/music/components/CoverArt';
import { getTrackCoverUrl } from '@/service/musicCover';
import { getMusicArtist, getMusicTitle } from '@/utils/music';
import type { MusicPlayedTrack } from '@/types/music';
import { useMusicPlayHistory } from './useMusicPlayHistory';
import styles from './MusicHomeScreen.module.css';

interface PlayHistorySectionProps {
    titleKey: string;
    emptyKey: string;
    playedTracks: MusicPlayedTrack[];
    isLoading: boolean;
    onPlay: (startIndex: number) => void;
}

const PlayHistorySection = ({
    titleKey,
    emptyKey,
    playedTracks,
    isLoading,
    onPlay,
}: PlayHistorySectionProps) => {
    const { t } = useI18n();

    return (
        <section className={styles.section}>
            <div className={styles.sectionHeader}>
                <h2 className={styles.sectionTitle}>{t(titleKey)}</h2>
            </div>

            <div className={styles.cardGrid}>
                {playedTracks.length === 0 ? (
                    <div className={styles.emptyCard}>{isLoading ? t('LOADING') : t(emptyKey)}</div>
                ) : (
                    playedTracks.map((playedTrack, index) => {
                        const trackTitle = getMusicTitle(playedTrack.track);
                        return (
                            <article key={playedTrack.track.id} className={styles.sectionCard}>
                                <div className={styles.cardCover}>
                                    <CoverArt
                                        src={getTrackCoverUrl(playedTrack.track.id, 256)}
                                        fallback={<Music size={40} opacity={0.5} />}
                                    />
                                </div>
                                <h3 className={styles.cardTitle}>{trackTitle}</h3>
                                <p className={styles.cardDescription}>
                                    {getMusicArtist(playedTrack.track)}
                                </p>
                                <span className={styles.metricCaption}>
                                    {t('MUSIC_HOME_PLAY_COUNT', {
                                        count: String(playedTrack.play_count ?? 0),
                                    })}
                                </span>
                                <div className={styles.actions}>
                                    <Button
                                        variant="contained"
                                        size="small"
                                        aria-label={t('MUSIC_HOME_PLAY_TRACK', {
                                            name: trackTitle,
                                        })}
                                        onClick={() => onPlay(index)}
                                    >
                                        {t('MUSIC_HOME_PLAY_NOW')}
                                    </Button>
                                </div>
                            </article>
                        );
                    })
                )}
            </div>
        </section>
    );
};

const MusicPlayHistorySections = () => {
    const {
        mostPlayedTracks,
        recentlyPlayedTracks,
        isLoadingMostPlayed,
        isLoadingRecentPlays,
        playMostPlayed,
        playRecentlyPlayed,
    } = useMusicPlayHistory();

    return (
        <>
            <PlayHistorySection
                titleKey="MUSIC_HOME_MOST_PLAYED"
                emptyKey="MUSIC_HOME_MOST_PLAYED_EMPTY"
                playedTracks={mostPlayedTracks}
                isLoading={isLoadingMostPlayed}
                onPlay={playMostPlayed}
            />
            <PlayHistorySection
                titleKey="MUSIC_HOME_RECENT_PLAYS"
                emptyKey="MUSIC_HOME_RECENT_PLAYS_EMPTY"
                playedTracks={recentlyPlayedTracks}
                isLoading={isLoadingRecentPlays}
                onPlay={playRecentlyPlayed}
            />
        </>
    );
};

export default MusicPlayHistorySections;
