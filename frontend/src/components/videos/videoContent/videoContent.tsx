import { useVideos } from '@/components/hooks/useVideos/useVideos';
import { VideoCatalogItemDto, VideoCatalogSectionDto } from '@/service/videoPlayback';
import { CircularProgress, IconButton, LinearProgress, Typography } from '@mui/material';
import { ChevronLeft, ChevronRight, Play, Videotape } from 'lucide-react';
import { useMemo, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import styles from './videoContent.module.css';

const statusLabel: Record<string, string> = {
	not_started: 'Nao iniciado',
	in_progress: 'Em andamento',
	completed: 'Assistido',
};

const sectionTitle: Record<string, string> = {
	continue: 'Continue Assistindo',
	series: 'Series',
	movies: 'Filmes',
	personal: 'Videos Pessoais',
	recent: 'Adicionados Recentemente',
};

const apiBase = `${import.meta.env.VITE_API_URL}/api/v1/files`;

const Thumbnail = ({ item }: { item: VideoCatalogItemDto }) => {
	return (
		<div className={styles.thumbnail}>
			<div className={styles.thumbFallback}>
				<Videotape size={38} color='white' />
			</div>
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
	);
};

const VideoRail = ({ section }: { section: VideoCatalogSectionDto }) => {
	const navigate = useNavigate();
	const railRef = useRef<HTMLDivElement>(null);

	if (section.items.length === 0) return null;

	const scrollBy = (direction: 1 | -1) => {
		const el = railRef.current;
		if (!el) return;
		el.scrollBy({ left: direction * 720, behavior: 'smooth' });
	};

	return (
		<section className={styles.railSection}>
			<div className={styles.railHeader}>
				<Typography variant='h5' className={styles.railTitle}>
					{sectionTitle[section.key] ?? section.title}
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
				{section.items.map((item) => (
					<button
						type='button'
						key={`${section.key}-${item.video.id}`}
						className={styles.videoCard}
						onClick={() => navigate(`/video/${item.video.id}`)}
					>
						<Thumbnail item={item} />
						<div className={styles.cardOverlay}>
							<div className={styles.cardTopLine}>
								<span className={styles.statusBadge}>{statusLabel[item.status] ?? statusLabel.not_started}</span>
								<span className={styles.formatTag}>{item.video.format.replace('.', '').toUpperCase()}</span>
							</div>
							<Typography className={styles.cardTitle}>{item.video.name}</Typography>
							<LinearProgress
								variant='determinate'
								value={item.progress_pct}
								className={styles.progress}
							/>
						</div>
						<div className={styles.playButtonWrap}>
							<Play size={18} />
						</div>
					</button>
				))}
			</div>
		</section>
	);
};

const Hero = ({ item }: { item: VideoCatalogItemDto }) => {
	const navigate = useNavigate();
	return (
		<section className={styles.hero}>
			<img
				className={styles.heroImage}
				src={`${apiBase}/video-thumbnail/${item.video.id}?width=1280&height=720`}
				alt={item.video.name}
			/>
			<div className={styles.heroShade} />
			<div className={styles.heroContent}>
				<p className={styles.heroEyebrow}>KuraNAS Video</p>
				<h1 className={styles.heroTitle}>{item.video.name}</h1>
				<p className={styles.heroMeta}>{statusLabel[item.status] ?? statusLabel.not_started}</p>
				<button type='button' className={styles.heroBtn} onClick={() => navigate(`/video/${item.video.id}`)}>
					<Play size={18} />
					<span>Assistir</span>
				</button>
			</div>
		</section>
	);
};

const VideoContent = () => {
	const { data, isLoading } = useVideos();

	const sections = data?.sections ?? [];
	const heroItem = useMemo(() => {
		const continueSection = sections.find((section) => section.key === 'continue');
		if (continueSection?.items?.[0]) return continueSection.items[0];
		const recentSection = sections.find((section) => section.key === 'recent');
		if (recentSection?.items?.[0]) return recentSection.items[0];
		return sections[0]?.items?.[0] ?? null;
	}, [sections]);

	if (isLoading) {
		return (
			<div className={styles.loadingState}>
				<CircularProgress size={44} />
				<Typography variant='h6'>Carregando catalogo de videos...</Typography>
			</div>
		);
	}

	if (!heroItem && sections.length === 0) {
		return (
			<div className={styles.emptyState}>
				<Videotape size={72} />
				<h2>Nenhum video encontrado</h2>
				<p>Atualize a biblioteca para montar seu catalogo.</p>
			</div>
		);
	}

	return (
		<div className={styles.page}>
			{heroItem && <Hero item={heroItem} />}
			<div className={styles.railsContainer}>
				{sections.map((section) => (
					<VideoRail key={section.key} section={section} />
				))}
			</div>
		</div>
	);
};

export default VideoContent;
