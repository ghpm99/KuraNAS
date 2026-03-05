import { formatSize } from '@/utils';
import { CircularProgress } from '@mui/material';
import { CalendarDays, ChevronLeft, ChevronRight, Expand, Info, Minus, Plus, Search, X } from 'lucide-react';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useImage, type IImageData } from '../hooks/imageProvider/imageProvider';
import { useIntersectionObserver } from '../hooks/IntersectionObserver/useIntersectionObserver';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import './imageContent.css';

const thumbnailWidth = 960;
const thumbnailHeight = 720;

const categoryLabels = {
	all: 'Todas',
	recent: 'Recentes',
	portrait: 'Retratos',
	landscape: 'Paisagens',
	screenshots: 'Capturas',
	camera: 'Câmera',
} as const;

type CategoryKey = keyof typeof categoryLabels;

const monthFormatter = new Intl.DateTimeFormat('pt-BR', {
	month: 'long',
	year: 'numeric',
});

const dateFormatter = new Intl.DateTimeFormat('pt-BR', {
	dateStyle: 'medium',
	timeStyle: 'short',
});

const imageDate = (image: IImageData): Date | null => {
	const candidates = [
		image.metadata?.datetime_original,
		image.metadata?.datetime,
		image.metadata?.createdAt,
		image.updated_at,
		image.created_at,
	];

	for (const candidate of candidates) {
		if (!candidate) {
			continue;
		}

		const parsed = new Date(candidate);
		if (!Number.isNaN(parsed.getTime())) {
			return parsed;
		}
	}

	return null;
};

const isRecent = (date: Date | null): boolean => {
	if (!date) {
		return false;
	}
	const last30days = Date.now() - 1000 * 60 * 60 * 24 * 30;
	return date.getTime() >= last30days;
};

const isScreenshot = (image: IImageData): boolean => {
	const sample = `${image.name} ${image.path}`.toLowerCase();
	return sample.includes('screenshot') || sample.includes('captura') || sample.includes('screen shot');
};

const imageCategoryCheck = (category: CategoryKey, image: IImageData, date: Date | null): boolean => {
	const width = image.metadata?.width ?? 0;
	const height = image.metadata?.height ?? 0;

	if (category === 'all') {
		return true;
	}
	if (category === 'recent') {
		return isRecent(date);
	}
	if (category === 'portrait') {
		return height > width;
	}
	if (category === 'landscape') {
		return width >= height;
	}
	if (category === 'screenshots') {
		return isScreenshot(image);
	}
	return Boolean(image.metadata?.make || image.metadata?.model);
};

const imageMetadataSummary = (image: IImageData): string => {
	const format = image.format ? `${image.format} - ` : '';
	return `${format}${formatSize(image.size)}`;
};

const thumbnailUrl = (id: number) =>
	`${getApiV1BaseUrl()}/files/thumbnail/${id}?width=${thumbnailWidth}&height=${thumbnailHeight}`;

const blobUrl = (id: number) => `${getApiV1BaseUrl()}/files/blob/${id}`;

const ImageContent = () => {
	const { images, fetchNextPage, hasNextPage, isFetchingNextPage } = useImage();
	const [activeCategory, setActiveCategory] = useState<CategoryKey>('all');
	const [search, setSearch] = useState('');
	const [viewerImageId, setViewerImageId] = useState<number | null>(null);
	const [zoom, setZoom] = useState(1);
	const [showDetails, setShowDetails] = useState(true);

	const closeViewer = useCallback(() => {
		setViewerImageId(null);
		setZoom(1);
	}, []);

	const imageDates = useMemo(() => {
		const entries = images.map((item) => [item.id, imageDate(item)] as const);
		return new Map<number, Date | null>(entries);
	}, [images]);

	const filteredImages = useMemo(() => {
		const searchValue = search.trim().toLowerCase();

		const filtered = images.filter((item) => {
			const date = imageDates.get(item.id) ?? null;
			if (!imageCategoryCheck(activeCategory, item, date)) {
				return false;
			}

			if (!searchValue) {
				return true;
			}

			const searchSample = [item.name, item.path, item.format, item.metadata?.make, item.metadata?.model]
				.filter(Boolean)
				.join(' ')
				.toLowerCase();

			return searchSample.includes(searchValue);
		});

		return filtered.sort((a, b) => {
			const aDate = imageDates.get(a.id)?.getTime() ?? 0;
			const bDate = imageDates.get(b.id)?.getTime() ?? 0;
			return bDate - aDate;
		});
	}, [images, imageDates, activeCategory, search]);

	const categoryCounts = useMemo(() => {
		const counts = {
			all: images.length,
			recent: 0,
			portrait: 0,
			landscape: 0,
			screenshots: 0,
			camera: 0,
		};

		for (const item of images) {
			const date = imageDates.get(item.id) ?? null;
			if (imageCategoryCheck('recent', item, date)) counts.recent += 1;
			if (imageCategoryCheck('portrait', item, date)) counts.portrait += 1;
			if (imageCategoryCheck('landscape', item, date)) counts.landscape += 1;
			if (imageCategoryCheck('screenshots', item, date)) counts.screenshots += 1;
			if (imageCategoryCheck('camera', item, date)) counts.camera += 1;
		}

		return counts;
	}, [images, imageDates]);

	const groupedImages = useMemo(() => {
		const grouped = new Map<string, { label: string; sortValue: number; items: IImageData[] }>();

		for (const item of filteredImages) {
			const date = imageDates.get(item.id) ?? null;
			const key = date ? `${date.getFullYear()}-${date.getMonth()}` : 'unknown';
			if (!grouped.has(key)) {
				grouped.set(key, {
					label: date ? monthFormatter.format(date) : 'Sem data registrada',
					sortValue: date?.getTime() ?? 0,
					items: [],
				});
			}
			grouped.get(key)?.items.push(item);
		}

		return Array.from(grouped.values()).sort((a, b) => b.sortValue - a.sortValue);
	}, [filteredImages, imageDates]);

	const activeIndex = useMemo(
		() => filteredImages.findIndex((image) => image.id === viewerImageId),
		[filteredImages, viewerImageId],
	);
	const activeImage = activeIndex >= 0 ? filteredImages[activeIndex] : null;
	const activeImageDate = activeImage ? (imageDates.get(activeImage.id) ?? null) : null;

	const openImage = (id: number) => {
		setViewerImageId(id);
		setZoom(1);
	};

	const goNext = useCallback(() => {
		if (filteredImages.length === 0 || activeIndex < 0) {
			return;
		}
		const nextImage = filteredImages[(activeIndex + 1) % filteredImages.length];
		if (!nextImage) {
			return;
		}
		setViewerImageId(nextImage.id);
		setZoom(1);
	}, [filteredImages, activeIndex]);

	const goPrevious = useCallback(() => {
		if (filteredImages.length === 0 || activeIndex < 0) {
			return;
		}
		const previous = activeIndex === 0 ? filteredImages.length - 1 : activeIndex - 1;
		const previousImage = filteredImages[previous];
		if (!previousImage) {
			return;
		}
		setViewerImageId(previousImage.id);
		setZoom(1);
	}, [filteredImages, activeIndex]);

	const increaseZoom = useCallback(() => {
		setZoom((value) => Math.min(5, Number((value + 0.2).toFixed(2))));
	}, []);

	const decreaseZoom = useCallback(() => {
		setZoom((value) => Math.max(0.5, Number((value - 0.2).toFixed(2))));
	}, []);

	const resetZoom = useCallback(() => {
		setZoom(1);
	}, []);

	useEffect(() => {
		if (!activeImage) {
			return;
		}

		const onKeyDown = (event: KeyboardEvent) => {
			if (event.key === 'Escape') closeViewer();
			if (event.key === 'ArrowRight') goNext();
			if (event.key === 'ArrowLeft') goPrevious();
			if (event.key === '+' || event.key === '=') increaseZoom();
			if (event.key === '-') decreaseZoom();
			if (event.key === '0') resetZoom();
			if (event.key.toLowerCase() === 'i') setShowDetails((value) => !value);
		};

		window.addEventListener('keydown', onKeyDown);
		return () => window.removeEventListener('keydown', onKeyDown);
	}, [activeImage, closeViewer, goNext, goPrevious, increaseZoom, decreaseZoom, resetZoom]);

	const { ref: loadMoreRef } = useIntersectionObserver<HTMLDivElement>({
		enabled: hasNextPage && !isFetchingNextPage,
		rootMargin: '500px',
		onIntersect: () => {
			if (hasNextPage && !isFetchingNextPage) {
				fetchNextPage();
			}
		},
	});

	return (
		<>
			<div className='images-toolbar'>
				<div>
					<h2>Galeria de fotos</h2>
					<p>
						{filteredImages.length} de {images.length} imagens
					</p>
				</div>
				<label className='images-search'>
					<Search size={16} />
					<input
						type='search'
						value={search}
						onChange={(event) => setSearch(event.target.value)}
						placeholder='Buscar por nome, pasta, câmera...'
					/>
				</label>
			</div>

			<div className='images-categories' role='tablist' aria-label='Categorias de imagens'>
				{(Object.keys(categoryLabels) as CategoryKey[]).map((key) => (
					<button
						key={key}
						type='button'
						className={`category-pill ${activeCategory === key ? 'is-active' : ''}`}
						onClick={() => setActiveCategory(key)}
					>
						<span>{categoryLabels[key]}</span>
						<strong>{categoryCounts[key]}</strong>
					</button>
				))}
			</div>

			{groupedImages.length === 0 && !isFetchingNextPage && (
				<div className='images-empty'>
					<h3>Nenhuma imagem encontrada</h3>
					<p>Tente remover filtros ou ajustar sua busca.</p>
				</div>
			)}

			<div className='images-sections'>
				{groupedImages.map((group) => (
					<section key={group.label} className='images-group'>
						<header>
							<CalendarDays size={16} />
							<h3>{group.label}</h3>
							<span>{group.items.length} fotos</span>
						</header>
						<div className='images-grid'>
							{group.items.map((item) => {
								const width = item.metadata?.width ?? 1;
								const height = item.metadata?.height ?? 1;
								const orientation = height > width ? 'portrait' : 'landscape';

								return (
									<button
										type='button'
										key={item.id}
										className={`photo-card ${orientation}`}
										onClick={() => openImage(item.id)}
										aria-label={`Abrir ${item.name}`}
									>
										<img className='thumbnail-img' src={thumbnailUrl(item.id)} alt={item.name} loading='lazy' />
										<div className='photo-overlay'>
											<p>{item.name}</p>
											<span>{imageMetadataSummary(item)}</span>
										</div>
									</button>
								);
							})}
						</div>
					</section>
				))}
			</div>

			<div ref={loadMoreRef} className='images-load-more-sentinel' aria-hidden='true' />

			{isFetchingNextPage && (
				<div className='loading-indicator'>
					<CircularProgress size={40} />
				</div>
			)}

			{!hasNextPage && images.length > 0 && <div className='end-message'>Todas as imagens carregadas</div>}

			{activeImage && (
				<div className='image-viewer-overlay' role='dialog' aria-modal='true'>
					<div className='image-viewer-topbar'>
						<div>
							<strong>{activeImage.name}</strong>
							<p>{activeImageDate ? dateFormatter.format(activeImageDate) : 'Data não disponível'}</p>
						</div>
						<div className='viewer-actions'>
							<button type='button' onClick={() => setShowDetails((value) => !value)} aria-label='Alternar detalhes'>
								<Info size={16} />
							</button>
							<button type='button' onClick={decreaseZoom} aria-label='Reduzir zoom'>
								<Minus size={16} />
							</button>
							<button type='button' onClick={resetZoom} aria-label='Resetar zoom'>
								<Expand size={16} />
							</button>
							<button type='button' onClick={increaseZoom} aria-label='Aumentar zoom'>
								<Plus size={16} />
							</button>
							<button type='button' onClick={closeViewer} aria-label='Fechar visualizador'>
								<X size={16} />
							</button>
						</div>
					</div>

					<div
						className='image-viewer-stage'
						onWheel={(event) => {
							event.preventDefault();
							if (event.deltaY < 0) increaseZoom();
							if (event.deltaY > 0) decreaseZoom();
						}}
					>
						<button type='button' className='viewer-nav left' onClick={goPrevious} aria-label='Imagem anterior'>
							<ChevronLeft size={24} />
						</button>
						<img
							src={blobUrl(activeImage.id)}
							alt={activeImage.name}
							className='viewer-image'
							style={{ transform: `scale(${zoom})` }}
						/>
						<button type='button' className='viewer-nav right' onClick={goNext} aria-label='Próxima imagem'>
							<ChevronRight size={24} />
						</button>
					</div>

					<div className='image-viewer-bottom'>
						<span>Zoom: {Math.round(zoom * 100)}%</span>
						<span>
							{activeIndex + 1} / {filteredImages.length}
						</span>
					</div>

					{showDetails && (
						<aside className='image-viewer-details'>
							<h4>Detalhes</h4>
							<p>
								<strong>Nome:</strong> {activeImage.name}
							</p>
							<p>
								<strong>Pasta:</strong> {activeImage.path}
							</p>
							<p>
								<strong>Formato:</strong> {activeImage.format || 'N/D'}
							</p>
							<p>
								<strong>Tamanho:</strong> {formatSize(activeImage.size)}
							</p>
							<p>
								<strong>Dimensões:</strong>{' '}
								{activeImage.metadata?.width && activeImage.metadata?.height
									? `${activeImage.metadata.width}x${activeImage.metadata.height}`
									: 'N/D'}
							</p>
							<p>
								<strong>Câmera:</strong> {activeImage.metadata?.make || activeImage.metadata?.model || 'N/D'}
							</p>
							<p>
								<strong>Lente:</strong> {activeImage.metadata?.lens_model || 'N/D'}
							</p>
							<p>
								<strong>ISO:</strong> {activeImage.metadata?.iso || 'N/D'}
							</p>
							<p>
								<strong>Focal:</strong> {activeImage.metadata?.focal_length || 'N/D'}
							</p>
							<p>
								<strong>Abertura:</strong> {activeImage.metadata?.f_number || 'N/D'}
							</p>
							<p>
								<strong>Exposição:</strong> {activeImage.metadata?.exposure_time || 'N/D'}
							</p>
						</aside>
					)}

					<div className='viewer-filmstrip'>
						{filteredImages.slice(Math.max(0, activeIndex - 8), activeIndex + 9).map((item) => (
							<button
								type='button'
								key={item.id}
								onClick={() => openImage(item.id)}
								className={`filmstrip-item ${activeImage.id === item.id ? 'is-active' : ''}`}
								aria-label={`Abrir ${item.name}`}
							>
								<img src={thumbnailUrl(item.id)} alt={item.name} loading='lazy' />
							</button>
						))}
					</div>
				</div>
			)}
		</>
	);
};

export default ImageContent;
