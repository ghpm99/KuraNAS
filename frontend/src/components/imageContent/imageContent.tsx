import { formatSize } from '@/utils';
import { CircularProgress } from '@mui/material';
import { CalendarDays, ChevronLeft, ChevronRight, Expand, Info, Minus, Plus, Search, X } from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useImage, type IImageData, type ImageGroupBy } from '../hooks/imageProvider/imageProvider';
import { useIntersectionObserver } from '../hooks/IntersectionObserver/useIntersectionObserver';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import useI18n from '@/components/i18n/provider/i18nContext';
import controlsStyles from './imageContentControls.module.css';
import './imageContent.css';

const thumbnailWidth = 960;
const thumbnailHeight = 720;

type CategoryKey = 'all' | 'recent' | 'portrait' | 'landscape' | 'screenshots' | 'camera';

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
	const { t } = useI18n();
	const { images, imageGroupBy, setImageGroupBy, fetchNextPage, hasNextPage, isFetchingNextPage } = useImage();
	const [activeCategory, setActiveCategory] = useState<CategoryKey>('all');
	const [search, setSearch] = useState('');
	const [viewerImageId, setViewerImageId] = useState<number | null>(null);
	const [zoom, setZoom] = useState(1);
	const [showDetails, setShowDetails] = useState(true);
	const isLoadingMoreRef = useRef(false);
	const locale = t('LOCALE');

	const categoryLabels: Record<CategoryKey, string> = useMemo(
		() => ({
			all: t('IMAGES_CATEGORY_ALL'),
			recent: t('IMAGES_CATEGORY_RECENT'),
			portrait: t('IMAGES_CATEGORY_PORTRAIT'),
			landscape: t('IMAGES_CATEGORY_LANDSCAPE'),
			screenshots: t('IMAGES_CATEGORY_SCREENSHOTS'),
			camera: t('IMAGES_CATEGORY_CAMERA'),
		}),
		[t],
	);
	const groupByLabels: Record<ImageGroupBy, string> = useMemo(
		() => ({
			date: t('IMAGES_GROUP_BY_DATE'),
			type: t('IMAGES_GROUP_BY_TYPE'),
			name: t('IMAGES_GROUP_BY_NAME'),
		}),
		[t],
	);
	const monthFormatter = useMemo(
		() =>
			new Intl.DateTimeFormat(locale, {
				month: 'long',
				year: 'numeric',
			}),
		[locale],
	);
	const dateFormatter = useMemo(
		() =>
			new Intl.DateTimeFormat(locale, {
				dateStyle: 'medium',
				timeStyle: 'short',
			}),
		[locale],
	);

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

		return filtered;
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
		const grouped = new Map<string, { label: string; items: IImageData[] }>();

		for (const item of filteredImages) {
			const date = imageDates.get(item.id) ?? null;
			const extension = item.format?.trim().toLowerCase() || t('IMAGES_GROUP_NO_FORMAT');
			const firstLetter = item.name.trim().charAt(0).toUpperCase() || '#';

			const key =
				imageGroupBy === 'date'
					? date
						? `${date.getFullYear()}-${date.getMonth()}`
						: 'unknown'
					: imageGroupBy === 'type'
						? extension
						: firstLetter;
			const label =
				imageGroupBy === 'date'
					? date
						? monthFormatter.format(date)
						: t('IMAGES_GROUP_NO_DATE')
					: imageGroupBy === 'type'
						? extension
						: t('IMAGES_GROUP_INITIAL', { letter: firstLetter });

			if (!grouped.has(key)) {
				grouped.set(key, {
					label,
					items: [],
				});
			}
			grouped.get(key)?.items.push(item);
		}

		return Array.from(grouped.values());
	}, [filteredImages, imageDates, imageGroupBy, monthFormatter, t]);

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

	const handleLoadMore = useCallback(async () => {
		if (!hasNextPage || isFetchingNextPage || isLoadingMoreRef.current) {
			return;
		}
		isLoadingMoreRef.current = true;
		try {
			await fetchNextPage();
		} finally {
			isLoadingMoreRef.current = false;
		}
	}, [fetchNextPage, hasNextPage, isFetchingNextPage]);

	const lastVisibleImageId = filteredImages.length > 0 ? filteredImages[filteredImages.length - 1]?.id : undefined;

	const { ref: loadMoreRef } = useIntersectionObserver<HTMLButtonElement>({
		enabled: hasNextPage && !isFetchingNextPage && filteredImages.length > 0,
		rootMargin: '500px',
		onIntersect: handleLoadMore,
	});

	return (
		<>
			<div className='images-toolbar'>
				<div>
					<h2>{t('IMAGES_TITLE')}</h2>
					<p>
						{t('IMAGES_COUNT_SUMMARY', { filtered: String(filteredImages.length), total: String(images.length) })}
					</p>
				</div>
				<label className='images-search'>
					<Search size={16} />
					<input
						type='search'
						value={search}
						onChange={(event) => setSearch(event.target.value)}
						placeholder={t('IMAGES_SEARCH_PLACEHOLDER')}
					/>
				</label>
				<label className={controlsStyles.groupingSelect}>
					<span>{t('IMAGES_GROUP_BY_LABEL')}</span>
					<select
						aria-label={t('IMAGES_GROUP_BY_ARIA')}
						value={imageGroupBy}
						onChange={(event) => setImageGroupBy(event.target.value as ImageGroupBy)}
					>
						{(Object.keys(groupByLabels) as ImageGroupBy[]).map((key) => (
							<option key={key} value={key}>
								{groupByLabels[key]}
							</option>
						))}
					</select>
				</label>
			</div>

			<div className='images-categories' role='tablist' aria-label={t('IMAGES_CATEGORIES_ARIA')}>
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
					<h3>{t('IMAGES_EMPTY_TITLE')}</h3>
					<p>{t('IMAGES_EMPTY_DESC')}</p>
				</div>
			)}

			<div className='images-sections'>
				{groupedImages.map((group) => (
					<section key={group.label} className='images-group'>
						<header>
							<CalendarDays size={16} />
							<h3>{group.label}</h3>
							<span>{t('IMAGES_PHOTOS_COUNT', { count: String(group.items.length) })}</span>
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
										ref={item.id === lastVisibleImageId ? loadMoreRef : undefined}
										className={`photo-card ${orientation}`}
										onClick={() => openImage(item.id)}
										aria-label={t('IMAGES_OPEN_IMAGE_ARIA', { name: item.name })}
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

			{isFetchingNextPage && (
				<div className='loading-indicator'>
					<CircularProgress size={40} />
				</div>
			)}

			{!hasNextPage && images.length > 0 && <div className='end-message'>{t('IMAGES_END_MESSAGE')}</div>}

			{activeImage && (
				<div className='image-viewer-overlay' role='dialog' aria-modal='true'>
					<div className='image-viewer-topbar'>
						<div>
							<strong>{activeImage.name}</strong>
							<p>{activeImageDate ? dateFormatter.format(activeImageDate) : t('IMAGES_DATE_UNAVAILABLE')}</p>
						</div>
						<div className='viewer-actions'>
							<button type='button' onClick={() => setShowDetails((value) => !value)} aria-label={t('IMAGES_TOGGLE_DETAILS')}>
								<Info size={16} />
							</button>
							<button type='button' onClick={decreaseZoom} aria-label={t('IMAGES_DECREASE_ZOOM')}>
								<Minus size={16} />
							</button>
							<button type='button' onClick={resetZoom} aria-label={t('IMAGES_RESET_ZOOM')}>
								<Expand size={16} />
							</button>
							<button type='button' onClick={increaseZoom} aria-label={t('IMAGES_INCREASE_ZOOM')}>
								<Plus size={16} />
							</button>
							<button type='button' onClick={closeViewer} aria-label={t('IMAGES_CLOSE_VIEWER')}>
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
						<button type='button' className='viewer-nav left' onClick={goPrevious} aria-label={t('IMAGES_PREVIOUS')}>
							<ChevronLeft size={24} />
						</button>
						<img
							src={blobUrl(activeImage.id)}
							alt={activeImage.name}
							className='viewer-image'
							style={{ transform: `scale(${zoom})` }}
						/>
						<button type='button' className='viewer-nav right' onClick={goNext} aria-label={t('IMAGES_NEXT')}>
							<ChevronRight size={24} />
						</button>
					</div>

					<div className='image-viewer-bottom'>
						<span>{t('IMAGES_ZOOM_LABEL')}: {Math.round(zoom * 100)}%</span>
						<span>
							{activeIndex + 1} / {filteredImages.length}
						</span>
					</div>

					{showDetails && (
						<aside className='image-viewer-details'>
							<h4>{t('IMAGES_DETAILS_TITLE')}</h4>
							<p>
								<strong>{t('IMAGES_DETAIL_NAME')}:</strong> {activeImage.name}
							</p>
							<p>
								<strong>{t('IMAGES_DETAIL_FOLDER')}:</strong> {activeImage.path}
							</p>
							<p>
								<strong>{t('IMAGES_DETAIL_FORMAT')}:</strong> {activeImage.format || t('COMMON_NOT_AVAILABLE')}
							</p>
							<p>
								<strong>{t('IMAGES_DETAIL_SIZE')}:</strong> {formatSize(activeImage.size)}
							</p>
							<p>
								<strong>{t('IMAGES_DETAIL_DIMENSIONS')}:</strong>{' '}
								{activeImage.metadata?.width && activeImage.metadata?.height
									? `${activeImage.metadata.width}x${activeImage.metadata.height}`
									: t('COMMON_NOT_AVAILABLE')}
							</p>
							<p>
								<strong>{t('IMAGES_DETAIL_CAMERA')}:</strong>{' '}
								{activeImage.metadata?.make || activeImage.metadata?.model || t('COMMON_NOT_AVAILABLE')}
							</p>
							<p>
								<strong>{t('IMAGES_DETAIL_LENS')}:</strong> {activeImage.metadata?.lens_model || t('COMMON_NOT_AVAILABLE')}
							</p>
							<p>
								<strong>{t('IMAGES_DETAIL_ISO')}:</strong> {activeImage.metadata?.iso || t('COMMON_NOT_AVAILABLE')}
							</p>
							<p>
								<strong>{t('IMAGES_DETAIL_FOCAL')}:</strong> {activeImage.metadata?.focal_length || t('COMMON_NOT_AVAILABLE')}
							</p>
							<p>
								<strong>{t('IMAGES_DETAIL_APERTURE')}:</strong> {activeImage.metadata?.f_number || t('COMMON_NOT_AVAILABLE')}
							</p>
							<p>
								<strong>{t('IMAGES_DETAIL_EXPOSURE')}:</strong>{' '}
								{activeImage.metadata?.exposure_time || t('COMMON_NOT_AVAILABLE')}
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
								aria-label={t('IMAGES_OPEN_IMAGE_ARIA', { name: item.name })}
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
