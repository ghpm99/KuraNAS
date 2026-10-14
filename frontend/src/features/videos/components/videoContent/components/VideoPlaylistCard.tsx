import { Typography } from '@mui/material';
import { EyeOff, Play, Videotape } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import {
    VideoPlaylistDto,
    getVideoPlaylistPosterUrl,
    getVideoPreviewUrl,
    getVideoThumbnailUrl,
} from '@/service/videoPlayback';
import { useHoverPreview } from '@/features/videos/hooks/useHoverPreview';
import { useLocalPoster } from '../useLocalPoster';
import { getVideoClassificationLabelKey } from '../videoClassificationLabel';
import styles from '../videoContent.module.css';

type VideoPlaylistCardProps = {
    playlist: VideoPlaylistDto;
    onSelect: (playlist: VideoPlaylistDto) => void;
    onPlay: (videoId: number, playlistId?: number | null) => void;
    onHide?: (playlist: VideoPlaylistDto) => void;
    focusVideoId?: number | null;
    badge?: string;
};

export default function VideoPlaylistCard({
    playlist,
    onSelect,
    onPlay,
    onHide,
    focusVideoId,
    badge,
}: VideoPlaylistCardProps) {
    const { t } = useI18n();
    const coverId = focusVideoId ?? playlist.cover_video_id;
    const coverThumb = coverId ? getVideoThumbnailUrl(coverId) : '';
    const { imageUrl, isPosterShown, handlePosterError } = useLocalPoster(
        getVideoPlaylistPosterUrl(playlist.id),
        coverThumb
    );
    const {
        isPreviewMounted,
        isPreviewLoaded,
        previewHandlers,
        handlePreviewLoad,
        handlePreviewError,
    } = useHoverPreview();

    return (
        <div className={styles.videoCardWrap}>
            <button
                type="button"
                className={styles.videoCard}
                onClick={() => onSelect(playlist)}
                {...previewHandlers}
            >
                <div
                    className={`${styles.thumbnail} ${isPreviewLoaded ? styles.thumbnailPreviewReady : ''}`}
                >
                    <div className={styles.thumbFallback}>
                        <Videotape size={38} color="white" />
                    </div>
                    {(coverId || isPosterShown) && (
                        <img
                            className={styles.thumbStatic}
                            loading="lazy"
                            src={imageUrl}
                            alt={playlist.name}
                            onError={isPosterShown ? handlePosterError : undefined}
                        />
                    )}
                    {coverId && isPreviewMounted && (
                        <img
                            className={styles.thumbPreview}
                            src={getVideoPreviewUrl(coverId)}
                            alt={t('VIDEO_PREVIEW_ALT', { name: playlist.name })}
                            onLoad={handlePreviewLoad}
                            onError={handlePreviewError}
                        />
                    )}
                </div>
                <div className={styles.cardOverlay}>
                    <div className={styles.cardTopLine}>
                        <span className={styles.statusBadge}>
                            {t('VIDEO_PLAYLIST_ITEM_COUNT', {
                                count: String(playlist.item_count),
                            })}
                        </span>
                        <span className={styles.formatTag}>
                            {t(getVideoClassificationLabelKey(playlist.classification))}
                        </span>
                    </div>
                    <Typography className={styles.cardTitle}>{playlist.name}</Typography>
                    {badge && <span className={styles.continueBadge}>{badge}</span>}
                </div>
            </button>
            <div className={styles.cardActionRow}>
                <button
                    type="button"
                    className={styles.playFromCardBtn}
                    onClick={() => onPlay(coverId ?? playlist.cover_video_id ?? 0, playlist.id)}
                    disabled={!coverId && !playlist.cover_video_id}
                >
                    <Play size={16} />
                    {t('VIDEO_PLAY')}
                </button>
                {onHide && (
                    <button
                        type="button"
                        className={styles.hideFromCardBtn}
                        aria-label={t('VIDEO_HIDE_PLAYLIST')}
                        title={t('VIDEO_HIDE_PLAYLIST')}
                        onClick={() => onHide(playlist)}
                    >
                        <EyeOff size={16} />
                    </button>
                )}
            </div>
        </div>
    );
}
