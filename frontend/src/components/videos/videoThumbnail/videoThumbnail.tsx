import type { VideoFileDto } from '@/service/videoPlayback';
import useI18n from '@/components/i18n/provider/i18nContext';
import { 
	Card, 
	CardContent, 
	Typography, 
	IconButton, 
	Box,
	Tooltip
} from '@mui/material';
import { Play, Videotape, Clock, Monitor } from 'lucide-react';
import './videoThumbnail.css';

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

	const formatFileSize = (bytes: number): string => {
		const sizes = ['B', 'KB', 'MB', 'GB'];
		if (bytes === 0) return '0 B';
		const i = Math.floor(Math.log(bytes) / Math.log(1024));
		return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
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
			className="video-thumbnail-card"
			onClick={handlePlay}
			onKeyDown={handleKeyDown}
			tabIndex={0}
			role="button"
			aria-label={t('VIDEO_PLAY_ARIA', { title: getVideoTitle() })}
		>
			{/* Thumbnail Container */}
			<Box className="thumbnail-container">
				<Box className="thumbnail-placeholder">
					<Videotape size={48} className="thumbnail-icon" />
					<IconButton 
						className="play-overlay"
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
			<CardContent className="video-info">
				<Typography 
					variant="subtitle2" 
					className="video-title"
					title={getVideoTitle()}
				>
					{getVideoTitle()}
				</Typography>

				<Box className="metadata-container">
					{/* Duration */}
					<Box className="metadata-item">
						<Clock size={14} className="metadata-icon" />
						<Typography variant="caption" className="metadata-text">
							{getVideoDuration()}
						</Typography>
					</Box>

					{/* Resolution */}
					<Box className="metadata-item">
						<Monitor size={14} className="metadata-icon" />
						<Typography variant="caption" className="metadata-text">
							{getVideoResolution()}
						</Typography>
					</Box>

					{/* Codec */}
					<Tooltip title={t('VIDEO_CODEC_TOOLTIP', { codec: getVideoCodec() })}>
						<Box className="metadata-item codec-badge">
							<Typography variant="caption" className="codec-text">
								{getVideoCodec()}
							</Typography>
						</Box>
					</Tooltip>

					{/* File Size */}
					<Typography variant="caption" className="file-size">
						{formatFileSize(video.size)}
					</Typography>
				</Box>
			</CardContent>
		</Card>
	);
};

export default VideoThumbnail;
