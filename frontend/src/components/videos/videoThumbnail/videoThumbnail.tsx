import type { VideoFileDto } from '@/service/videoPlayback';
import useI18n from '@/components/i18n/provider/i18nContext';
import { formatSize } from '@/utils';
import {
	Card,
	CardContent,
	Typography,
	IconButton,
	Box,
	Tooltip
} from '@mui/material';
import { Play, Videotape, Clock, Monitor } from 'lucide-react';
import styles from './videoThumbnail.module.css';

interface VideoThumbnailProps {
	video: VideoFileDto;
	onPlay: (video: VideoFileDto) => void;
}

const VideoThumbnail = ({ video, onPlay }: VideoThumbnailProps) => {
	const { t } = useI18n();
	const getVideoTitle = (): string => {
		if (video.metadata?.format_name) {
			return video.metadata.format_name;
		}
		return video.name;
	};

	const getVideoDuration = (): string => {
		if (video.metadata?.duration) {
			return video.metadata.duration;
		}
		return t('VIDEO_UNKNOWN');
	};

	const getVideoResolution = (): string => {
		if (video.metadata?.width && video.metadata?.height) {
			return `${video.metadata.width}x${video.metadata.height}`;
		}
		return t('VIDEO_UNKNOWN');
	};

	const getVideoCodec = (): string => {
		if (video.metadata?.codec_name) {
			return video.metadata.codec_name.toUpperCase();
		}
		return t('VIDEO_UNKNOWN');
	};

	const handlePlay = () => {
		onPlay(video);
	};

	const handleKeyDown = (event: React.KeyboardEvent) => {
		if (event.key === 'Enter' || event.key === ' ') {
			event.preventDefault();
			handlePlay();
		}
	};

	return (
		<Card
			className={styles.card}
			onClick={handlePlay}
			onKeyDown={handleKeyDown}
			tabIndex={0}
			role="button"
			aria-label={t('VIDEO_PLAY_ARIA', { title: getVideoTitle() })}
		>
			{/* Thumbnail Container */}
			<Box className={styles.thumbnailContainer}>
				<Box className={styles.thumbnailPlaceholder}>
					<Videotape size={48} className={styles.thumbnailIcon} />
					<IconButton
						className={styles.playOverlay}
						onClick={(e) => {
							e.stopPropagation();
							handlePlay();
						}}
						size="large"
					>
						<Play size={24} />
					</IconButton>
				</Box>
			</Box>

			{/* Video Info */}
			<CardContent className={styles.videoInfo}>
				<Typography
					variant="subtitle2"
					className={styles.videoTitle}
					title={getVideoTitle()}
				>
					{getVideoTitle()}
				</Typography>

				<Box className={styles.metadataContainer}>
					{/* Duration */}
					<Box className={styles.metadataItem}>
						<Clock size={14} className={styles.metadataIcon} />
						<Typography variant="caption" className={styles.metadataText}>
							{getVideoDuration()}
						</Typography>
					</Box>

					{/* Resolution */}
					<Box className={styles.metadataItem}>
						<Monitor size={14} className={styles.metadataIcon} />
						<Typography variant="caption" className={styles.metadataText}>
							{getVideoResolution()}
						</Typography>
					</Box>

					{/* Codec */}
					<Tooltip title={t('VIDEO_CODEC_TOOLTIP', { codec: getVideoCodec() })}>
						<Box className={`${styles.metadataItem} ${styles.codecBadge}`}>
							<Typography variant="caption" className={styles.codecText}>
								{getVideoCodec()}
							</Typography>
						</Box>
					</Tooltip>

					{/* File Size */}
					<Typography variant="caption" className={styles.fileSize}>
						{formatSize(video.size)}
					</Typography>
				</Box>
			</CardContent>
		</Card>
	);
};

export default VideoThumbnail;
