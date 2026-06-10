import { Button } from '@mui/material';
import { type ReactNode } from 'react';
import { Link } from 'react-router-dom';
import useI18n from '@/components/i18n/provider/i18nContext';
import { type VideoPlaylistDto } from '@/service/videoPlayback';
import VideoPlaylistCard from './VideoPlaylistCard';
import styles from '../videoContent.module.css';

type VideoSectionPlaylistGridProps = {
    titleKey: string;
    descriptionKey: string;
    emptyKey: string;
    playlists: VideoPlaylistDto[];
    onSelectPlaylist: (playlist: VideoPlaylistDto) => void;
    onPlayVideo: (videoId: number, playlistId?: number | null) => void;
    badge?: string;
    action?: ReactNode;
};

export default function VideoSectionPlaylistGrid({
    titleKey,
    descriptionKey,
    emptyKey,
    playlists,
    onSelectPlaylist,
    onPlayVideo,
    badge,
    action,
}: VideoSectionPlaylistGridProps) {
    const { t } = useI18n();

    return (
        <section className={styles.sectionBlock}>
            <div className={styles.sectionHeaderRow}>
                <div className={styles.sectionHeader}>
                    <h2>{t(titleKey)}</h2>
                    <p>{t(descriptionKey)}</p>
                </div>
                {action}
            </div>
            {playlists.length === 0 ? (
                <div className={styles.sectionEmpty}>{t(emptyKey)}</div>
            ) : (
                <div className={styles.gridCards}>
                    {playlists.map((playlist) => (
                        <VideoPlaylistCard
                            key={playlist.id}
                            playlist={playlist}
                            onSelect={onSelectPlaylist}
                            onPlay={onPlayVideo}
                            badge={badge}
                        />
                    ))}
                </div>
            )}
        </section>
    );
}

export const VideoSectionActionLink = ({ to }: { to: string }) => {
    const { t } = useI18n();

    return (
        <Button component={Link} to={to} variant="text" size="small">
            {t('VIDEO_OPEN_SECTION')}
        </Button>
    );
};
