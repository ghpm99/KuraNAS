import {
	useAllVideoFiles,
	useVideoPlaylistDetail,
	useVideoPlaylists,
} from '@/components/hooks/useVideos/useVideos';
import {
	addVideoToPlaylist,
	getVideoPlaylistById,
	removeVideoFromPlaylist,
	reorderVideoPlaylist,
	updateVideoPlaylistName,
	VideoFileDto,
	VideoPlaylistDto,
	getVideoPlaybackState,
} from '@/service/videoPlayback';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Alert, CircularProgress, Snackbar, TextField, Typography } from '@mui/material';
import { ArrowDown, ArrowLeft, ArrowUp, Play, Plus, Trash2, Videotape } from 'lucide-react';
import { useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import styles from './videoContent.module.css';

const classificationTitle: Record<string, string> = {
	anime: 'Animes',
	series: 'Series',
	movie: 'Filmes',
	personal: 'Gravacoes Pessoais',
	program: 'Videos de Programas',
};

const slugify = (value: string) =>
	value
		.normalize('NFD')
		.replace(/[\u0300-\u036f]/g, '')
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-+|-+$/g, '');

const apiBase = `${import.meta.env.VITE_API_URL}/api/v1/files`;

const PlaylistCard = ({
	playlist,
	onSelect,
	onPlay,
	focusVideoId,
	badge,
}: {
	playlist: VideoPlaylistDto;
	onSelect: (playlist: VideoPlaylistDto) => void;
	onPlay: (videoId: number, playlistId?: number | null) => void;
	focusVideoId?: number | null;
	badge?: string;
}) => {
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
							<img className={styles.thumbPreview} loading='lazy' src={coverPreview} alt={`${playlist.name} preview`} />
						</>
					)}
				</div>
				<div className={styles.cardOverlay}>
					<div className={styles.cardTopLine}>
						<span className={styles.statusBadge}>{playlist.item_count} videos</span>
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
				Reproduzir
			</button>
		</div>
	);
};

const PlaylistDetail = ({
	playlist,
	onChanged,
}: {
	playlist: VideoPlaylistDto;
	onChanged: () => void;
}) => {
	const navigate = useNavigate();
	const [searchParams, setSearchParams] = useSearchParams();
	const queryClient = useQueryClient();
	const coverId = playlist.cover_video_id || playlist.items[0]?.video.id;
	const playlistSlug = searchParams.get('playlist') || slugify(playlist.name);
	const selectedVideoQuery = searchParams.get('video');
	const selectedVideoID = selectedVideoQuery ? Number(selectedVideoQuery) : null;
	const [nameDraft, setNameDraft] = useState(playlist.name);

	const orderedItems = useMemo(
		() => [...playlist.items].sort((a, b) => a.order_index - b.order_index || a.id - b.id),
		[playlist.items],
	);

	const refresh = async () => {
		await Promise.all([
			queryClient.invalidateQueries({ queryKey: ['video-playlists'] }),
			queryClient.invalidateQueries({ queryKey: ['video-playlist', playlist.id] }),
			queryClient.invalidateQueries({ queryKey: ['video-home-catalog'] }),
		]);
		onChanged();
	};

	const renameMutation = useMutation({
		mutationFn: async () => updateVideoPlaylistName(playlist.id, nameDraft),
		onSuccess: refresh,
	});

	const removeMutation = useMutation({
		mutationFn: async (videoId: number) => removeVideoFromPlaylist(playlist.id, videoId),
		onSuccess: refresh,
	});

	const reorderMutation = useMutation({
		mutationFn: async (items: { video_id: number; order_index: number }[]) => reorderVideoPlaylist(playlist.id, items),
		onSuccess: refresh,
	});

	const moveItem = (index: number, direction: -1 | 1) => {
		const target = index + direction;
		if (target < 0 || target >= orderedItems.length) return;
		const swapped = [...orderedItems];
		const current = swapped[index];
		const next = swapped[target];
		if (!current || !next) return;
		swapped[index] = next;
		swapped[target] = current;
		reorderMutation.mutate(
			swapped.map((item, idx) => ({
				video_id: item.video.id,
				order_index: idx,
			})),
		);
	};

	return (
		<div className={styles.page}>
			<section className={styles.hero}>
				{coverId && (
					<img
						className={styles.heroImage}
						src={`${apiBase}/video-thumbnail/${coverId}?width=1280&height=720`}
						alt={playlist.name}
					/>
				)}
				<div className={styles.heroShade} />
				<div className={styles.heroContent}>
					<button type='button' className={styles.backBtn} onClick={() => setSearchParams({})}>
						<ArrowLeft size={16} />
						<span>Voltar para videos</span>
					</button>
					<p className={styles.heroEyebrow}>{(playlist.classification || 'personal').toUpperCase()}</p>
					<h1 className={styles.heroTitle}>{playlist.name}</h1>
					<p className={styles.heroMeta}>{playlist.item_count} videos nesta playlist</p>
				</div>
			</section>

			<section className={styles.managementPanel}>
				<div className={styles.panelHeader}>
					<Typography variant='h6'>Editar playlist</Typography>
				</div>
				<div className={styles.renameRow}>
					<TextField
						size='small'
						value={nameDraft}
						onChange={(event) => setNameDraft(event.target.value)}
						placeholder='Nome de exibicao'
					/>
					<button
						type='button'
						className={styles.actionBtn}
						onClick={() => renameMutation.mutate()}
						disabled={renameMutation.isPending || !nameDraft.trim()}
					>
						Salvar nome
					</button>
				</div>
			</section>

			<section className={styles.detailList}>
				{orderedItems.map((item, index) => (
					<div
						key={item.video.id}
						className={`${styles.detailItem} ${selectedVideoID === item.video.id ? styles.detailItemSelected : ''}`}
					>
						<button
							type='button'
							className={styles.detailMainButton}
							onClick={() => {
								const currentPoint = `/videos?playlist=${encodeURIComponent(playlistSlug)}&video=${item.video.id}`;
								setSearchParams({ playlist: playlistSlug, video: String(item.video.id) });
								navigate(`/video/${item.video.id}`, {
									state: { from: currentPoint, playlistId: playlist.id },
								});
							}}
						>
							<div className={styles.detailThumb}>
								<img
									loading='lazy'
									src={`${apiBase}/video-thumbnail/${item.video.id}?width=320&height=180`}
									alt={item.video.name}
								/>
							</div>
							<div className={styles.detailMeta}>
								<Typography className={styles.detailTitle}>{item.video.name}</Typography>
								<Typography className={styles.detailSub}>
									{item.video.format.toUpperCase()} · {item.source_kind === 'manual' ? 'manual' : 'auto'}
								</Typography>
							</div>
							<div className={styles.detailPlay}>
								<Play size={16} />
							</div>
						</button>
						<div className={styles.itemActions}>
							<button
								type='button'
								className={styles.iconBtn}
								onClick={() => moveItem(index, -1)}
								disabled={index === 0 || reorderMutation.isPending}
							>
								<ArrowUp size={14} />
							</button>
							<button
								type='button'
								className={styles.iconBtn}
								onClick={() => moveItem(index, 1)}
								disabled={index === orderedItems.length - 1 || reorderMutation.isPending}
							>
								<ArrowDown size={14} />
							</button>
							<button
								type='button'
								className={styles.iconBtnDanger}
								onClick={() => removeMutation.mutate(item.video.id)}
								disabled={removeMutation.isPending}
							>
								<Trash2 size={14} />
							</button>
						</div>
					</div>
				))}
			</section>
		</div>
	);
};

const VideoContent = () => {
	const [searchParams, setSearchParams] = useSearchParams();
	const navigate = useNavigate();
	const queryClient = useQueryClient();
	const { data: playlists = [], isLoading: isLoadingPlaylists } = useVideoPlaylists();
	const { data: allVideos = [], isLoading: isLoadingVideos } = useAllVideoFiles();
	const { data: playbackState } = useQuery({
		queryKey: ['video-playback-state'],
		queryFn: getVideoPlaybackState,
		retry: false,
	});
	const [videoSearch, setVideoSearch] = useState('');
	const [selectedPlaylistPerVideo, setSelectedPlaylistPerVideo] = useState<Record<number, number>>({});
	const [feedback, setFeedback] = useState<{ open: boolean; message: string; severity: 'success' | 'error' }>({
		open: false,
		message: '',
		severity: 'success',
	});

	const playlistSlug = searchParams.get('playlist') || '';
	const selectedPlaylistSummary = useMemo(() => {
		if (!playlistSlug) return null;
		return playlists.find((playlist) => slugify(playlist.name) === playlistSlug) ?? null;
	}, [playlistSlug, playlists]);

	const { data: selectedPlaylistDetail, isLoading: isLoadingPlaylist } = useVideoPlaylistDetail(selectedPlaylistSummary?.id);

	const continuePlaylists = useMemo(() => {
		return [...playlists]
			.filter((playlist) => Boolean(playlist.last_played_at))
			.sort((a, b) => {
				const aTime = a.last_played_at ? new Date(a.last_played_at).getTime() : 0;
				const bTime = b.last_played_at ? new Date(b.last_played_at).getTime() : 0;
				return bTime - aTime;
			});
	}, [playlists]);

	const groupedPlaylists = useMemo(() => {
		const result: Record<string, VideoPlaylistDto[]> = {
			anime: [],
			series: [],
			movie: [],
			personal: [],
			program: [],
		};
		for (const playlist of playlists) {
			const key = playlist.classification || 'personal';
			if (!result[key]) result[key] = [];
			result[key].push(playlist);
		}
		return result;
	}, [playlists]);

	const filteredVideos = useMemo(() => {
		if (!videoSearch.trim()) return allVideos;
		const query = videoSearch.toLowerCase();
		return allVideos.filter(
			(video) =>
				video.name.toLowerCase().includes(query) ||
				video.parent_path.toLowerCase().includes(query) ||
				video.format.toLowerCase().includes(query),
		);
	}, [allVideos, videoSearch]);

	const { data: playlistMembershipMap = {} } = useQuery({
		queryKey: ['video-playlist-membership', playlists.map((playlist) => playlist.id).join(',')],
		enabled: playlists.length > 0,
		queryFn: async () => {
			const entries = await Promise.all(
				playlists.map(async (playlist) => {
					const detail = await getVideoPlaylistById(playlist.id);
					return [playlist.id, new Set(detail.items.map((item) => item.video.id))] as const;
				}),
			);
			return Object.fromEntries(entries) as Record<number, Set<number>>;
		},
	});

	const addToPlaylistMutation = useMutation({
		mutationFn: async ({ playlistId, videoId }: { playlistId: number; videoId: number }) =>
			addVideoToPlaylist(playlistId, videoId),
		onSuccess: async () => {
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: ['video-playlists'] }),
				queryClient.invalidateQueries({ queryKey: ['video-playlist'] }),
				queryClient.invalidateQueries({ queryKey: ['video-playlist-membership'] }),
				queryClient.invalidateQueries({ queryKey: ['video-home-catalog'] }),
			]);
			setFeedback({
				open: true,
				message: 'Video adicionado a playlist com sucesso.',
				severity: 'success',
			});
		},
		onError: () => {
			setFeedback({
				open: true,
				message: 'Nao foi possivel adicionar o video a playlist.',
				severity: 'error',
			});
		},
	});

	const onSelectPlaylist = (playlist: VideoPlaylistDto) => {
		setSearchParams({ playlist: slugify(playlist.name) });
	};

	const playVideo = (videoId: number, playlistId?: number | null) => {
		if (!videoId) return;
		const from = `/videos${window.location.search}`;
		navigate(`/video/${videoId}`, { state: { from, playlistId: playlistId ?? null } });
	};

	if (isLoadingPlaylists || isLoadingVideos) {
		return (
			<div className={styles.loadingState}>
				<CircularProgress size={44} />
				<Typography variant='h6'>Carregando videos...</Typography>
			</div>
		);
	}

	if (selectedPlaylistSummary) {
		if (isLoadingPlaylist || !selectedPlaylistDetail) {
			return (
				<div className={styles.loadingState}>
					<CircularProgress size={40} />
					<Typography variant='h6'>Carregando playlist...</Typography>
				</div>
			);
		}
		return <PlaylistDetail playlist={selectedPlaylistDetail} onChanged={() => undefined} />;
	}

	return (
		<div className={styles.page}>
			<section className={styles.sectionBlock}>
				<div className={styles.sectionHeader}>
					<h2>Continuar assistindo</h2>
					<p>Playlists com reproducao recente.</p>
				</div>
				{continuePlaylists.length === 0 ? (
					<div className={styles.sectionEmpty}>Nenhuma playlist com reproducao recente.</div>
				) : (
					<div className={styles.gridCards}>
						{continuePlaylists.map((playlist) => {
							const isCurrent = playbackState?.playback_state.playlist_id === playlist.id;
							const focusVideoId = isCurrent ? playbackState?.playback_state.video_id : playlist.cover_video_id;
							const badge = isCurrent ? 'Em andamento' : 'Retomar';
							return (
								<PlaylistCard
									key={`continue-${playlist.id}`}
									playlist={playlist}
									onSelect={onSelectPlaylist}
									onPlay={playVideo}
									focusVideoId={focusVideoId}
									badge={badge}
								/>
							);
						})}
					</div>
				)}
			</section>

			<section className={styles.sectionBlock}>
				<div className={styles.sectionHeader}>
					<h2>Playlists</h2>
					<p>Catálogo organizado automaticamente.</p>
				</div>
				{Object.entries(groupedPlaylists).map(([key, list]) => {
					if (list.length === 0) return null;
					return (
						<div key={key} className={styles.groupBlock}>
							<h3>{classificationTitle[key] ?? key}</h3>
							<div className={styles.gridCards}>
								{list.map((playlist) => (
									<PlaylistCard
										key={playlist.id}
										playlist={playlist}
										onSelect={onSelectPlaylist}
										onPlay={playVideo}
									/>
								))}
							</div>
						</div>
					);
				})}
			</section>

			<section className={styles.sectionBlock}>
				<div className={styles.sectionHeader}>
					<h2>Todos</h2>
					<p>Todos os videos do sistema com acao rapida para adicionar a playlists.</p>
				</div>
				<div className={styles.searchRow}>
					<TextField
						size='small'
						fullWidth
						placeholder='Buscar video por nome, pasta ou formato'
						value={videoSearch}
						onChange={(event) => setVideoSearch(event.target.value)}
					/>
				</div>
				<div className={styles.allVideosList}>
					{filteredVideos.map((video: VideoFileDto) => {
						const selectedPlaylist = selectedPlaylistPerVideo[video.id] ?? playlists[0]?.id;
						const isAlreadyInPlaylist = Boolean(
							selectedPlaylist &&
								playlistMembershipMap[selectedPlaylist] &&
								playlistMembershipMap[selectedPlaylist]?.has(video.id),
						);
						return (
							<div className={styles.allVideoItem} key={video.id}>
								<div className={styles.allVideoThumb}>
									<img
										loading='lazy'
										src={`${apiBase}/video-thumbnail/${video.id}?width=240&height=135`}
										alt={video.name}
									/>
								</div>
								<div className={styles.allVideoMeta}>
									<h4>{video.name}</h4>
									<p>
										{video.parent_path} · {video.format.toUpperCase()}
									</p>
								</div>
								<div className={styles.allVideoActions}>
									<button type='button' className={styles.actionBtn} onClick={() => playVideo(video.id, null)}>
										<Play size={14} />
										Reproduzir
									</button>
									<select
										className={styles.playlistSelect}
										value={selectedPlaylist ?? ''}
										onChange={(event) =>
											setSelectedPlaylistPerVideo((prev) => ({
												...prev,
												[video.id]: Number(event.target.value),
											}))
										}
									>
										{playlists.map((playlist) => (
											<option key={`add-${video.id}-${playlist.id}`} value={playlist.id}>
												{playlist.name}
											</option>
										))}
									</select>
									<button
										type='button'
										className={styles.actionBtn}
										disabled={!selectedPlaylist || addToPlaylistMutation.isPending || isAlreadyInPlaylist}
										onClick={() => {
											if (!selectedPlaylist) return;
											if (isAlreadyInPlaylist) {
												setFeedback({
													open: true,
													message: 'Esse video ja esta na playlist selecionada.',
													severity: 'error',
												});
												return;
											}
											addToPlaylistMutation.mutate({ playlistId: selectedPlaylist, videoId: video.id });
										}}
									>
										<Plus size={14} />
										{isAlreadyInPlaylist ? 'Ja adicionado' : 'Adicionar'}
									</button>
								</div>
							</div>
						);
					})}
				</div>
			</section>
			<Snackbar
				open={feedback.open}
				autoHideDuration={2600}
				onClose={() => setFeedback((prev) => ({ ...prev, open: false }))}
				anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
			>
				<Alert
					severity={feedback.severity}
					variant='filled'
					onClose={() => setFeedback((prev) => ({ ...prev, open: false }))}
				>
					{feedback.message}
				</Alert>
			</Snackbar>
		</div>
	);
};

export default VideoContent;
