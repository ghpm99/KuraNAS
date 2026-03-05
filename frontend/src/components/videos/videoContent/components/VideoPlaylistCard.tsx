import { Typography } from '@mui/material';
import { Play, Videotape } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { VideoPlaylistDto } from '@/service/videoPlayback';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import styles from '../videoContent.module.css';

const apiBase = `${getApiV1BaseUrl()}/files`;

type VideoPlaylistCardProps = {
	playlist: VideoPlaylistDto;
	onSelect: (playlist: VideoPlaylistDto) => void;
	onPlay: (videoId: number, playlistId?: number | null) => void;
	focusVideoId?: number | null;
	badge?: string;
};

export default function VideoPlaylistCard({ playlist, onSelect, onPlay, focusVideoId, badge }: VideoPlaylistCardProps) {
	const { t } = useI18n();
	const coverId = focusVideoId ?? playlist.cover_video_id;
	const coverThumb = coverId ? `${apiBase}/video-thumbnail/${coverId}?width=640&height=360` : '';
	const coverPreview = coverId ? `${apiBase}/video-preview/${coverId}?width=640&height=360` : '';

	return (
		<div className={styles.videoCardWrap}>
			<button type='button' className={styles.videoCard} onClick={() => onSelect(playlist)}>
				<div className={styles.thumbnail}>
					<div className={styles.thumbFallback}>
						<Videotape size={38} color='white' />
					</div>
					{coverId && (
						<>
							<img className={styles.thumbStatic} loading='lazy' src={coverThumb} alt={playlist.name} />
							<img
								className={styles.thumbPreview}
								loading='lazy'
								src={coverPreview}
								alt={t('VIDEO_PREVIEW_ALT', { name: playlist.name })}
							/>
						</>
					)}
				</div>
				<div className={styles.cardOverlay}>
					<div className={styles.cardTopLine}>
						<span className={styles.statusBadge}>{t('VIDEO_PLAYLIST_ITEM_COUNT', { count: String(playlist.item_count) })}</span>
						<span className={styles.formatTag}>{(playlist.classification || 'personal').toUpperCase()}</span>
					</div>
					<Typography className={styles.cardTitle}>{playlist.name}</Typography>
					{badge && <span className={styles.continueBadge}>{badge}</span>}
				</div>
			</button>
			<button
				type='button'
				className={styles.playFromCardBtn}
				onClick={() => onPlay(coverId ?? playlist.cover_video_id ?? 0, playlist.id)}
				disabled={!coverId && !playlist.cover_video_id}
			>
				<Play size={16} />
				{t('VIDEO_PLAY')}
			</button>
		</div>
	);
}
