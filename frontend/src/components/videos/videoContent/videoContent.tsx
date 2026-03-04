import {
	useAllVideoFiles,
	useVideoPlaylistDetail,
	useVideoPlaylists,
	useVideosWithoutPlaylist,
} from '@/components/hooks/useVideos/useVideos';
import { VideoFileDto, VideoPlaylistDto } from '@/service/videoPlayback';
import { CircularProgress, IconButton, Typography } from '@mui/material';
import { ArrowLeft, ChevronLeft, ChevronRight, Play, Videotape } from 'lucide-react';
import { useMemo, useRef } from 'react';
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

const PlaylistDetail = ({ playlist }: { playlist: VideoPlaylistDto }) => {
	const navigate = useNavigate();
	const [searchParams, setSearchParams] = useSearchParams();
	const coverId = playlist.cover_video_id || playlist.items[0]?.video.id;
	const playlistSlug = searchParams.get('playlist') || slugify(playlist.name);
	const selectedVideoQuery = searchParams.get('video');
	const selectedVideoID = selectedVideoQuery ? Number(selectedVideoQuery) : null;

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

			<section className={styles.detailList}>
				{playlist.items.map((item) => (
					<button
						type='button'
						key={item.video.id}
						className={`${styles.detailItem} ${selectedVideoID === item.video.id ? styles.detailItemSelected : ''}`}
						onClick={() => {
							const currentPoint = `/videos?playlist=${encodeURIComponent(playlistSlug)}&video=${item.video.id}`;
							setSearchParams({ playlist: playlistSlug, video: String(item.video.id) });
							navigate(`/video/${item.video.id}`, { state: { from: currentPoint } });
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
				))}
			</section>
		</div>
	);
};

const VideoContent = () => {
	const [searchParams, setSearchParams] = useSearchParams();
	const { data: playlists = [], isLoading } = useVideoPlaylists();
	const { data: allVideos = [], isLoading: isLoadingAllVideos } = useAllVideoFiles();
	const { data: unassignedVideos = [], isLoading: isLoadingUnassigned } = useVideosWithoutPlaylist();

	const virtualPlaylists = useMemo(() => {
		const all = createVirtualPlaylist('Todos', 'todos', allVideos);
		const unassigned = createVirtualPlaylist('Sem playlist', 'sem-playlist', unassignedVideos);
		return [all, unassigned].filter((p) => p.item_count > 0);
	}, [allVideos, unassignedVideos]);

	const mergedPlaylists = useMemo(() => [...virtualPlaylists, ...playlists], [playlists, virtualPlaylists]);

	const playlistSlug = searchParams.get('playlist') || '';
	const selectedPlaylistSummary = useMemo(() => {
		if (!playlistSlug) return null;
		return mergedPlaylists.find((playlist) => slugify(playlist.name) === playlistSlug) ?? null;
	}, [playlistSlug, mergedPlaylists]);

	const isVirtualSelected = Boolean(selectedPlaylistSummary && selectedPlaylistSummary.id === -1);
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
		return <PlaylistDetail playlist={isVirtualSelected ? selectedPlaylistSummary : (selectedPlaylistDetail as VideoPlaylistDto)} />;
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
				{Object.entries(grouped).map(([key, list]) => (
					<PlaylistRail key={key} title={classificationTitle[key] ?? key} playlists={list} onSelect={onSelectPlaylist} />
				))}
			</div>
		</div>
	);
};

export default VideoContent;
