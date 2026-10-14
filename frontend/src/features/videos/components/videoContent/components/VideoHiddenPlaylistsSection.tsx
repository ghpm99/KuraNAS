import { Eye } from 'lucide-react';
import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';
import useI18n from '@/components/i18n/provider/i18nContext';
import { type VideoPlaylistDto } from '@/service/videoPlayback';
import styles from '../videoContent.module.css';

type VideoHiddenPlaylistsSectionProps = {
    playlists: VideoPlaylistDto[];
    isChangingVisibility?: boolean;
    hasNextPage?: boolean;
    isFetchingNextPage?: boolean;
    fetchNextPage?: () => void;
    onShowPlaylist?: (playlist: VideoPlaylistDto) => void;
};

export default function VideoHiddenPlaylistsSection({
    playlists,
    isChangingVisibility = false,
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
    onShowPlaylist,
}: VideoHiddenPlaylistsSectionProps) {
    const { t } = useI18n();

    return (
        <section className={styles.sectionBlock}>
            <div className={styles.sectionHeader}>
                <h2>{t('VIDEO_SECTION_HIDDEN')}</h2>
                <p>{t('VIDEO_SECTION_HIDDEN_DESCRIPTION')}</p>
            </div>
            {playlists.length === 0 ? (
                <div className={styles.sectionEmpty}>{t('VIDEO_SECTION_HIDDEN_EMPTY')}</div>
            ) : (
                <div className={styles.hiddenPlaylistList}>
                    {playlists.map((playlist) => (
                        <div key={playlist.id} className={styles.hiddenPlaylistRow}>
                            <div>
                                <strong>{playlist.name}</strong>
                                <div className={styles.detailSub}>
                                    {t('VIDEO_PLAYLIST_ITEM_COUNT', {
                                        count: String(playlist.item_count),
                                    })}
                                </div>
                            </div>
                            {onShowPlaylist && (
                                <button
                                    type="button"
                                    className={styles.actionBtn}
                                    disabled={isChangingVisibility}
                                    onClick={() => onShowPlaylist(playlist)}
                                >
                                    <Eye size={14} />
                                    {t('VIDEO_SHOW_PLAYLIST')}
                                </button>
                            )}
                        </div>
                    ))}
                </div>
            )}
            <LoadMoreSentinel
                hasNextPage={hasNextPage}
                isFetchingNextPage={isFetchingNextPage}
                fetchNextPage={fetchNextPage}
            />
        </section>
    );
}
