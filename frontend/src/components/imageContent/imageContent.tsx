import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useImage, type IImageData, type ImageGroupBy } from '@/components/providers/imageProvider/imageProvider';
import { useIntersectionObserver } from '@/components/hooks/IntersectionObserver/useIntersectionObserver';
import { useImageViewer } from '@/components/hooks/useImageViewer/useImageViewer';
import useI18n from '@/components/i18n/provider/i18nContext';
import { useSearchParams } from 'react-router-dom';
import ImageCategoryTabs from './components/ImageCategoryTabs';
import ImageGroupsGrid from './components/ImageGroupsGrid';
import ImageToolbar from './components/ImageToolbar';
import ImageViewerModal from './components/ImageViewerModal';
import './imageContent.css';

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
		if (!candidate) continue;
		const parsed = new Date(candidate);
		if (!Number.isNaN(parsed.getTime())) return parsed;
	}
	return null;
};

const isRecent = (date: Date | null): boolean => {
	if (!date) return false;
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
	if (category === 'all') return true;
	if (category === 'recent') return isRecent(date);
	if (category === 'portrait') return height > width;
	if (category === 'landscape') return width >= height;
	if (category === 'screenshots') return isScreenshot(image);
	return Boolean(image.metadata?.make || image.metadata?.model);
};

export default function ImageContent() {
	const { t } = useI18n();
	const [searchParams, setSearchParams] = useSearchParams();
	const { images, imageGroupBy, setImageGroupBy, fetchNextPage, hasNextPage, isFetchingNextPage } = useImage();
	const [activeCategory, setActiveCategory] = useState<CategoryKey>('all');
	const [search, setSearch] = useState('');
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

	const imageDates = useMemo(() => new Map<number, Date | null>(images.map((item) => [item.id, imageDate(item)] as const)), [images]);

	const filteredImages = useMemo(() => {
		const searchValue = search.trim().toLowerCase();
		return images.filter((item) => {
			const date = imageDates.get(item.id) ?? null;
			if (!imageCategoryCheck(activeCategory, item, date)) return false;
			if (!searchValue) return true;
			const searchSample = [item.name, item.path, item.format, item.metadata?.make, item.metadata?.model]
				.filter(Boolean)
				.join(' ')
				.toLowerCase();
			return searchSample.includes(searchValue);
		});
	}, [images, imageDates, activeCategory, search]);

	const categoryCounts = useMemo(() => {
		const counts: Record<CategoryKey, number> = {
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
			if (!grouped.has(key)) grouped.set(key, { label, items: [] });
			grouped.get(key)?.items.push(item);
		}
		return Array.from(grouped.values());
	}, [filteredImages, imageDates, imageGroupBy, monthFormatter, t]);

	const {
		activeImage,
		activeIndex,
		zoom,
		showDetails,
		setShowDetails,
		openImage,
		closeViewer,
		goNext,
		goPrevious,
		increaseZoom,
		decreaseZoom,
		resetZoom,
	} = useImageViewer(filteredImages);

	const requestedImageId = Number(searchParams.get('image') ?? '');
	const activeImageDate = activeImage ? (imageDates.get(activeImage.id) ?? null) : null;

	const updateImageSearchParam = useCallback(
		(imageId: number | null) => {
			const nextSearchParams = new URLSearchParams(searchParams);
			if (imageId == null) {
				nextSearchParams.delete('image');
			} else {
				nextSearchParams.set('image', String(imageId));
			}
			setSearchParams(nextSearchParams);
		},
		[searchParams, setSearchParams],
	);

	const handleOpenImage = useCallback(
		(imageId: number) => {
			openImage(imageId);
			updateImageSearchParam(imageId);
		},
		[openImage, updateImageSearchParam],
	);

	const handleCloseViewer = useCallback(() => {
		closeViewer();
		updateImageSearchParam(null);
	}, [closeViewer, updateImageSearchParam]);

	useEffect(() => {
		if (!Number.isFinite(requestedImageId) || requestedImageId <= 0) {
			return;
		}

		const hasRequestedImage = filteredImages.some((image) => image.id === requestedImageId);
		if (hasRequestedImage) {
			openImage(requestedImageId);
		}
	}, [filteredImages, openImage, requestedImageId]);

	const handleLoadMore = useCallback(async () => {
		if (!hasNextPage || isFetchingNextPage || isLoadingMoreRef.current) return;
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
			<ImageToolbar
				filteredCount={filteredImages.length}
				totalCount={images.length}
				search={search}
				groupBy={imageGroupBy}
				groupByLabels={groupByLabels}
				onSearchChange={setSearch}
				onGroupByChange={setImageGroupBy}
			/>
			<ImageCategoryTabs
				activeCategory={activeCategory}
				labels={categoryLabels}
				counts={categoryCounts}
				onSelect={setActiveCategory}
			/>
			<ImageGroupsGrid
				groups={groupedImages}
				totalImages={images.length}
				isFetchingNextPage={isFetchingNextPage}
				hasNextPage={Boolean(hasNextPage)}
				lastVisibleImageId={lastVisibleImageId}
				loadMoreRef={loadMoreRef}
				onOpenImage={handleOpenImage}
			/>
			{activeImage && (
				<ImageViewerModal
					activeImage={activeImage}
					activeIndex={activeIndex}
					activeImageDate={activeImageDate}
					dateFormatter={dateFormatter}
					filteredImages={filteredImages}
					zoom={zoom}
					showDetails={showDetails}
					onToggleDetails={() => setShowDetails((value) => !value)}
					onDecreaseZoom={decreaseZoom}
					onResetZoom={resetZoom}
					onIncreaseZoom={increaseZoom}
					onClose={handleCloseViewer}
					onPrevious={goPrevious}
					onNext={goNext}
					onOpenImage={handleOpenImage}
				/>
			)}
		</>
	);
}
