import { useVideos } from '@/components/hooks/useVideos/useVideos';
import { useVideoPlayer } from '@/components/hooks/videoPlayerProvider/videoPlayerProvider';
import { Box, Card, CardContent, CircularProgress, Grid, IconButton, Typography } from '@mui/material';
import { Play, Videotape } from 'lucide-react';
import { useEffect, useRef } from 'react';
import './videoContent.module.css';
import styles from './videoContent.module.css';
import { IVideoData } from '@/types/video';

const VideoContent = () => {
	const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isLoading } = useVideos();

	const { playVideo, playlist, setPlaylist } = useVideoPlayer();
	const lastItemRef = useRef<HTMLDivElement>(null);

	const videos = data?.pages.flatMap((page) => page.data) || [];

	useEffect(() => {
		if (videos.length > 0) {
			//setPlaylist(videos);
		}
	}, [videos.length, setPlaylist, videos]);

	useEffect(() => {
		if (!lastItemRef.current || !hasNextPage || isFetchingNextPage) return;

		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0].isIntersecting) {
					fetchNextPage();
				}
			},
			{ threshold: 0.1 },
		);

		observer.observe(lastItemRef.current);

		return () => observer.disconnect();
	}, [fetchNextPage, hasNextPage, isFetchingNextPage]);

	const getVideoTitle = (video: IVideoData): string => {
		return video.name;
	};

	const getVideoDuration = (video: IVideoData): string => {
		if (video.metadata?.duration) {
			return video.metadata.duration;
		}
		return 'Unknown';
	};

	const getVideoResolution = (video: IVideoData): string => {
		if (video.metadata?.width && video.metadata?.height) {
			return `${video.metadata.width}x${video.metadata.height}`;
		}
		return 'Unknown';
	};

	const formatFileSize = (bytes: number): string => {
		const sizes = ['B', 'KB', 'MB', 'GB'];
		if (bytes === 0) return '0 B';
		const i = Math.floor(Math.log(bytes) / Math.log(1024));
		return Math.round((bytes / Math.pow(1024, i)) * 100) / 100 + ' ' + sizes[i];
	};

	if (isLoading) {
		return (
			<div className='video-content'>
				<div className='loading-container'>
					<CircularProgress size={40} />
					<Typography variant='h6' sx={{ mt: 2 }}>
						Carregando vídeos...
					</Typography>
				</div>
			</div>
		);
	}

	return (
		<>
			<div className={styles.title}>
				<Typography variant='h4' sx={{ mb: 3, fontWeight: 'bold' }}>
					Vídeos
				</Typography>
			</div>
			<div className={styles['video-content']}>
				<Grid container spacing={3}>
					{videos.map((video, index) => {
						const isLastItem = index === videos.length - 1;
						return (
							<Grid size={{ xs: 12, sm: 6, md: 4, lg: 3 }} key={video.id} ref={isLastItem ? lastItemRef : null}>
								<Card
									className={styles['video-card']}
									onClick={() => playVideo(video)}
									sx={{
										cursor: 'pointer',
										transition: 'transform 0.2s, box-shadow 0.2s',
										'&:hover': {
											transform: 'translateY(-4px)',
											boxShadow: 4,
										},
									}}
								>
									<Box className={styles['video-thumbnail']}>
										<Videotape size={48} color='white' />
										<IconButton
											className={styles['play-button']}
											size='small'
											onClick={(e) => {
												e.stopPropagation();
												playVideo(video);
											}}
										>
											<Play size={20} color='white' />
										</IconButton>
									</Box>

									<CardContent sx={{ p: 2 }}>
										<Typography variant='subtitle2' noWrap sx={{ fontWeight: 'medium', mb: 1 }}>
											{getVideoTitle(video)}
										</Typography>

										<Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.5 }}>
											<Typography variant='caption' color='text.secondary'>
												Duração: {getVideoDuration(video)}
											</Typography>
											<Typography variant='caption' color='text.secondary'>
												Resolução: {getVideoResolution(video)}
											</Typography>
											<Typography variant='caption' color='text.secondary'>
												Tamanho: {formatFileSize(video.size)}
											</Typography>
										</Box>
									</CardContent>
								</Card>
							</Grid>
						);
					})}
				</Grid>

				{isFetchingNextPage && (
					<div className={styles['loading-indicator']}>
						<CircularProgress size={40} />
					</div>
				)}

				{!hasNextPage && videos.length > 0 && (
					<div className={styles['end-message']}>
						<Typography variant='body2' color='text.secondary'>
							Todos os vídeos foram carregados
						</Typography>
					</div>
				)}

				{videos.length === 0 && !isLoading && (
					<div className={styles['empty-state']}>
						<Videotape size={64} color='text.secondary' />
						<Typography variant='h6' color='text.secondary' sx={{ mt: 2 }}>
							Nenhum vídeo encontrado
						</Typography>
					</div>
				)}
			</div>
		</>
	);
};

export default VideoContent;
