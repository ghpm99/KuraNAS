import { VideoPlaylistDto } from '@/service/videoPlayback';
import useI18n from '@/components/i18n/provider/i18nContext';
import VideoPlaylistCard from './VideoPlaylistCard';
import styles from '../videoContent.module.css';

type VideoCatalogSectionsProps = {
    continuePlaylists: VideoPlaylistDto[];
    groupedPlaylists: {
        classificationTitle: Record<string, string>;
        grouped: Record<string, VideoPlaylistDto[]>;
    };
    onSelectPlaylist: (playlist: VideoPlaylistDto) => void;
    onPlayVideo: (videoId: number, playlistId?: number | null) => void;
};

export default function VideoCatalogSections({
    continuePlaylists,
    groupedPlaylists,
    onSelectPlaylist,
    onPlayVideo,
}: VideoCatalogSectionsProps) {
    const { t } = useI18n();

    return (
        <>
            <section className={styles.sectionBlock}>
                <div className={styles.sectionHeader}>
                    <h2>{t('VIDEO_CONTINUE_WATCHING')}</h2>
                    <p>{t('VIDEO_RECENT_PLAYLISTS_DESC')}</p>
                </div>
                {continuePlaylists.length === 0 ? (
                    <div className={styles.sectionEmpty}>{t('VIDEO_NO_RECENT_PLAYLISTS')}</div>
                ) : (
                    <div className={styles.gridCards}>
                        {continuePlaylists.map((playlist) => (
                            <VideoPlaylistCard
                                key={`continue-${playlist.id}`}
                                playlist={playlist}
                                onSelect={onSelectPlaylist}
                                onPlay={onPlayVideo}
                                badge={t('VIDEO_CONTINUE_BADGE_RESUME')}
                            />
                        ))}
                    </div>
                )}
            </section>

            <section className={styles.sectionBlock}>
                <div className={styles.sectionHeader}>
                    <h2>{t('VIDEO_PLAYLISTS')}</h2>
                    <p>{t('VIDEO_PLAYLISTS_DESC')}</p>
                </div>
                {Object.entries(groupedPlaylists.grouped).map(([key, list]) => {
                    if (list.length === 0) return null;
                    return (
                        <div key={key} className={styles.groupBlock}>
                            <h3>{groupedPlaylists.classificationTitle[key] ?? key}</h3>
                            <div className={styles.gridCards}>
                                {list.map((playlist) => (
                                    <VideoPlaylistCard
                                        key={playlist.id}
                                        playlist={playlist}
                                        onSelect={onSelectPlaylist}
                                        onPlay={onPlayVideo}
                                    />
                                ))}
                            </div>
                        </div>
                    );
                })}
            </section>
        </>
    );
}
