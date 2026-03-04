import {
	useAllVideoFiles,
	useVideoHomeCatalog,
	useVideoPlaylistDetail,
	useVideoPlaylists,
	useVideosWithoutPlaylist,
} from '@/components/hooks/useVideos/useVideos';
import {
	addVideoToPlaylist,
	removeVideoFromPlaylist,
	reorderVideoPlaylist,
	updateVideoPlaylistName,
	VideoCatalogItemDto,
	VideoFileDto,
	VideoPlaylistDto,
} from '@/service/videoPlayback';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { CircularProgress, IconButton, TextField, Typography } from '@mui/material';
import {
	ArrowDown,
	ArrowLeft,
	ArrowUp,
	ChevronLeft,
	ChevronRight,
	Play,
	Plus,
	Trash2,
	Videotape,
} from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import styles from './videoContent.module.css';

const classificationTitle: Record<string, string> = {
	virtual: 'Visao Geral',
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

const createVirtualPlaylist = (name: string, slug: string, videos: VideoFileDto[]): VideoPlaylistDto => ({
	id: -1,
	type: 'custom',
	source_path: `virtual:${slug}`,
	name,
	is_hidden: false,
	is_auto: false,
	group_mode: 'single',
	classification: 'personal',
	item_count: videos.length,
	cover_video_id: videos[0]?.id ?? null,
	created_at: new Date().toISOString(),
	updated_at: new Date().toISOString(),
	last_played_at: null,
	items: videos.map((video, index) => ({
		id: -(index + 1),
		order_index: index,
		source_kind: 'manual',
		status: 'not_started',
		video,
	})),
});

const PlaylistCard = ({
	playlist,
	onSelect,
}: {
	playlist: VideoPlaylistDto;
	onSelect: (playlist: VideoPlaylistDto) => void;
}) => {
	const coverId = playlist.cover_video_id;
	const coverThumb = coverId ? `${apiBase}/video-thumbnail/${coverId}?width=640&height=360` : '';
	const coverPreview = coverId ? `${apiBase}/video-preview/${coverId}?width=640&height=360` : '';

	return (
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
			</div>
			<div className={styles.playButtonWrap}>
				<Play size={18} />
			</div>
		</button>
	);
};

const ContinueCard = ({
	item,
	onPlay,
}: {
	item: VideoCatalogItemDto;
	onPlay: (videoId: number) => void;
}) => {
	const pct = Math.max(0, Math.min(100, item.progress_pct || 0));
	return (
		<button type='button' className={styles.videoCard} onClick={() => onPlay(item.video.id)}>
			<div className={styles.thumbnail}>
				<img
					className={styles.thumbStatic}
					loading='lazy'
					src={`${apiBase}/video-thumbnail/${item.video.id}?width=640&height=360`}
					alt={item.video.name}
				/>
				<img
					className={styles.thumbPreview}
					loading='lazy'
					src={`${apiBase}/video-preview/${item.video.id}?width=640&height=360`}
					alt={`${item.video.name} preview`}
				/>
			</div>
			<div className={styles.cardOverlay}>
				<div className={styles.cardTopLine}>
					<span className={styles.statusBadge}>Continuar</span>
					<span className={styles.formatTag}>{Math.round(pct)}%</span>
				</div>
				<Typography className={styles.cardTitle}>{item.video.name}</Typography>
				<div className={styles.progressTrack}>
					<div className={styles.progressFill} style={{ width: `${pct}%` }} />
				</div>
			</div>
		</button>
	);
};

const PlaylistRail = ({
	title,
	playlists,
	onSelect,
}: {
	title: string;
	playlists: VideoPlaylistDto[];
	onSelect: (playlist: VideoPlaylistDto) => void;
}) => {
	const railRef = useRef<HTMLDivElement>(null);

	if (playlists.length === 0) return null;

	const scrollBy = (direction: 1 | -1) => {
		railRef.current?.scrollBy({ left: direction * 720, behavior: 'smooth' });
	};

	return (
		<section className={styles.railSection}>
			<div className={styles.railHeader}>
				<Typography variant='h5' className={styles.railTitle}>
					{title}
				</Typography>
				<div className={styles.railActions}>
					<IconButton className={styles.railNavBtn} onClick={() => scrollBy(-1)}>
						<ChevronLeft size={18} />
					</IconButton>
					<IconButton className={styles.railNavBtn} onClick={() => scrollBy(1)}>
						<ChevronRight size={18} />
					</IconButton>
				</div>
			</div>
			<div className={styles.rail} ref={railRef}>
				{playlists.map((playlist) => (
					<PlaylistCard key={`${playlist.source_path}-${playlist.name}`} playlist={playlist} onSelect={onSelect} />
				))}
			</div>
		</section>
	);
};

const ContinueRail = ({
	items,
	onPlay,
}: {
	items: VideoCatalogItemDto[];
	onPlay: (videoId: number) => void;
}) => {
	const railRef = useRef<HTMLDivElement>(null);
	if (items.length === 0) return null;

	const scrollBy = (direction: 1 | -1) => {
		railRef.current?.scrollBy({ left: direction * 720, behavior: 'smooth' });
	};

	return (
		<section className={styles.railSection}>
			<div className={styles.railHeader}>
				<Typography variant='h5' className={styles.railTitle}>
					Continuar assistindo
				</Typography>
				<div className={styles.railActions}>
					<IconButton className={styles.railNavBtn} onClick={() => scrollBy(-1)}>
						<ChevronLeft size={18} />
					</IconButton>
					<IconButton className={styles.railNavBtn} onClick={() => scrollBy(1)}>
						<ChevronRight size={18} />
					</IconButton>
				</div>
			</div>
			<div className={styles.rail} ref={railRef}>
				{items.map((item) => (
					<ContinueCard key={item.video.id} item={item} onPlay={onPlay} />
				))}
			</div>
		</section>
	);
};

const PlaylistDetail = ({
	playlist,
	allVideos,
	onChanged,
}: {
	playlist: VideoPlaylistDto;
	allVideos: VideoFileDto[];
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

	useEffect(() => {
		setNameDraft(playlist.name);
	}, [playlist.name]);

	const isVirtual = playlist.id < 0;
	const orderedItems = useMemo(
		() => [...playlist.items].sort((a, b) => a.order_index - b.order_index || a.id - b.id),
		[playlist.items],
	);

	const existingVideoIds = useMemo(() => new Set(orderedItems.map((item) => item.video.id)), [orderedItems]);
	const candidates = useMemo(
		() => allVideos.filter((video) => !existingVideoIds.has(video.id)).slice(0, 120),
		[allVideos, existingVideoIds],
	);

	const refresh = async () => {
		await Promise.all([
			queryClient.invalidateQueries({ queryKey: ['video-playlists'] }),
			queryClient.invalidateQueries({ queryKey: ['video-playlist', playlist.id] }),
			queryClient.invalidateQueries({ queryKey: ['video-unassigned'] }),
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

	const addMutation = useMutation({
		mutationFn: async (videoId: number) => addVideoToPlaylist(playlist.id, videoId),
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
						<span>Voltar para catalogo</span>
					</button>
					<p className={styles.heroEyebrow}>{(playlist.classification || 'personal').toUpperCase()}</p>
					<h1 className={styles.heroTitle}>{playlist.name}</h1>
					<p className={styles.heroMeta}>{playlist.item_count} videos nesta playlist</p>
				</div>
			</section>

			{!isVirtual && (
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
			)}

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
									state: { from: currentPoint, playlistId: playlist.id > 0 ? playlist.id : null },
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
									{item.video.format.toUpperCase()} · {item.source_kind === 'manual' ? 'adicionado manualmente' : 'auto'}
								</Typography>
							</div>
							<div className={styles.detailPlay}>
								<Play size={16} />
							</div>
						</button>
						{!isVirtual && (
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
						)}
					</div>
				))}
			</section>

			{!isVirtual && candidates.length > 0 && (
				<section className={styles.managementPanel}>
					<div className={styles.panelHeader}>
						<Typography variant='h6'>Adicionar videos</Typography>
					</div>
					<div className={styles.addGrid}>
						{candidates.map((video) => (
							<button
								type='button'
								key={video.id}
								className={styles.addItemBtn}
								onClick={() => addMutation.mutate(video.id)}
								disabled={addMutation.isPending}
							>
								<span className={styles.addItemText}>{video.name}</span>
								<Plus size={14} />
							</button>
						))}
					</div>
				</section>
			)}
		</div>
	);
};

const VideoContent = () => {
	const [searchParams, setSearchParams] = useSearchParams();
	const navigate = useNavigate();
	const { data: playlists = [], isLoading } = useVideoPlaylists();
	const { data: allVideos = [], isLoading: isLoadingAllVideos } = useAllVideoFiles();
	const { data: unassignedVideos = [], isLoading: isLoadingUnassigned } = useVideosWithoutPlaylist();
	const { data: homeCatalog } = useVideoHomeCatalog();

	const continueItems = useMemo(
		() => homeCatalog?.sections.find((section) => section.key === 'continue')?.items ?? [],
		[homeCatalog],
	);

	const allCatalogVideos = useMemo(() => {
		const map = new Map<number, VideoFileDto>();
		for (const video of allVideos) {
			map.set(video.id, video);
		}
		for (const section of homeCatalog?.sections ?? []) {
			for (const item of section.items) {
				if (!map.has(item.video.id)) {
					map.set(item.video.id, item.video);
				}
			}
		}
		return Array.from(map.values());
	}, [allVideos, homeCatalog]);

	const virtualPlaylists = useMemo(() => {
		const all = createVirtualPlaylist('Todos', 'todos', allCatalogVideos);
		const unassigned = createVirtualPlaylist('Sem playlist', 'sem-playlist', unassignedVideos);
		return [all, unassigned].filter((playlist) => playlist.item_count > 0);
	}, [allCatalogVideos, unassignedVideos]);

	const mergedPlaylists = useMemo(() => [...virtualPlaylists, ...playlists], [playlists, virtualPlaylists]);

	const playlistSlug = searchParams.get('playlist') || '';
	const selectedPlaylistSummary = useMemo(() => {
		if (!playlistSlug) return null;
		return mergedPlaylists.find((playlist) => slugify(playlist.name) === playlistSlug) ?? null;
	}, [playlistSlug, mergedPlaylists]);

	const isVirtualSelected = Boolean(selectedPlaylistSummary && selectedPlaylistSummary.id < 0);
	const selectedPlaylistId = !isVirtualSelected ? selectedPlaylistSummary?.id : undefined;
	const { data: selectedPlaylistDetail, isLoading: isLoadingPlaylist } = useVideoPlaylistDetail(selectedPlaylistId);

	const grouped = useMemo(() => {
		const result: Record<string, VideoPlaylistDto[]> = {
			virtual: [...virtualPlaylists],
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
	}, [playlists, virtualPlaylists]);

	const heroPlaylist = mergedPlaylists[0] ?? null;

	const onSelectPlaylist = (playlist: VideoPlaylistDto) => {
		setSearchParams({ playlist: slugify(playlist.name) });
	};

	const playFromVideos = (videoId: number, playlistId?: number | null) => {
		const from = `/videos${window.location.search}`;
		navigate(`/video/${videoId}`, { state: { from, playlistId: playlistId ?? null } });
	};

	if (isLoading || isLoadingAllVideos || isLoadingUnassigned) {
		return (
			<div className={styles.loadingState}>
				<CircularProgress size={44} />
				<Typography variant='h6'>Carregando playlists de videos...</Typography>
			</div>
		);
	}

	if (selectedPlaylistSummary) {
		if (!isVirtualSelected && (isLoadingPlaylist || !selectedPlaylistDetail)) {
			return (
				<div className={styles.loadingState}>
					<CircularProgress size={40} />
					<Typography variant='h6'>Carregando playlist...</Typography>
				</div>
			);
		}
		return (
			<PlaylistDetail
				playlist={isVirtualSelected ? selectedPlaylistSummary : (selectedPlaylistDetail as VideoPlaylistDto)}
				allVideos={allCatalogVideos}
				onChanged={() => {
					// no-op: mutations already invalidate queries
				}}
			/>
		);
	}

	if (mergedPlaylists.length === 0) {
		return (
			<div className={styles.emptyState}>
				<Videotape size={72} />
				<h2>Nenhum video encontrado</h2>
				<p>Escaneie os arquivos para gerar playlists automaticas.</p>
			</div>
		);
	}

	return (
		<div className={styles.page}>
			{heroPlaylist && (
				<section className={styles.hero}>
					{heroPlaylist.cover_video_id && (
						<img
							className={styles.heroImage}
							src={`${apiBase}/video-thumbnail/${heroPlaylist.cover_video_id}?width=1280&height=720`}
							alt={heroPlaylist.name}
						/>
					)}
					<div className={styles.heroShade} />
					<div className={styles.heroContent}>
						<p className={styles.heroEyebrow}>KuraNAS Video</p>
						<h1 className={styles.heroTitle}>Explore suas playlists inteligentes</h1>
						<p className={styles.heroMeta}>
							Clique em qualquer playlist para abrir em /videos?playlist=... e compartilhar o mesmo ponto por URL.
						</p>
					</div>
				</section>
			)}

			<div className={styles.railsContainer}>
				<ContinueRail items={continueItems} onPlay={(videoId) => playFromVideos(videoId, null)} />
				{Object.entries(grouped).map(([key, list]) => (
					<PlaylistRail key={key} title={classificationTitle[key] ?? key} playlists={list} onSelect={onSelectPlaylist} />
				))}
			</div>
		</div>
	);
};

export default VideoContent;
