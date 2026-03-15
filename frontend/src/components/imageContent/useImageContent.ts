import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useSearchParams, useLocation } from 'react-router-dom';
import { getImageSectionFromPath } from '@/components/images/navigation';
import { useImage, type ImageGroupBy } from '@/components/providers/imageProvider/imageProvider';
import { useIntersectionObserver } from '@/components/hooks/IntersectionObserver/useIntersectionObserver';
import { useImageViewer } from '@/components/hooks/useImageViewer/useImageViewer';
import useI18n from '@/components/i18n/provider/i18nContext';
import {
	buildAutomaticAlbumCollections,
	buildFolderCollections,
	getCollectionTitleFromPath,
	getImageDate,
	matchesImageSearch,
	matchesImageSection,
	type AutomaticAlbumKey,
} from './imageLibraryData';

const automaticAlbumTranslationKeys: Record<AutomaticAlbumKey, { titleKey: string; descriptionKey: string }> = {
	travel: {
		titleKey: 'IMAGES_ALBUM_TRAVEL',
		descriptionKey: 'IMAGES_ALBUM_TRAVEL_DESCRIPTION',
	},
	documents: {
		titleKey: 'IMAGES_ALBUM_DOCUMENTS',
		descriptionKey: 'IMAGES_ALBUM_DOCUMENTS_DESCRIPTION',
	},
	wallpapers: {
		titleKey: 'IMAGES_ALBUM_WALLPAPERS',
		descriptionKey: 'IMAGES_ALBUM_WALLPAPERS_DESCRIPTION',
	},
	memes: {
		titleKey: 'IMAGES_ALBUM_MEMES',
		descriptionKey: 'IMAGES_ALBUM_MEMES_DESCRIPTION',
	},
	others: {
		titleKey: 'IMAGES_ALBUM_OTHERS',
		descriptionKey: 'IMAGES_ALBUM_OTHERS_DESCRIPTION',
	},
};

const selectionSearchParamMap = {
	folders: 'folder',
	albums: 'album',
} as const;

export const useImageContent = () => {
	const { t } = useI18n();
	const location = useLocation();
	const [searchParams, setSearchParams] = useSearchParams();
	const { images, status, imageGroupBy, setImageGroupBy, fetchNextPage, hasNextPage, isFetchingNextPage } = useImage();
	const [search, setSearch] = useState('');
	const isLoadingMoreRef = useRef(false);
	const locale = t('LOCALE');
	const activeSection = getImageSectionFromPath(location.pathname);

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

	const imageDates = useMemo(() => new Map<number, Date | null>(images.map((item) => [item.id, getImageDate(item)] as const)), [images]);
	const folderCollections = useMemo(() => buildFolderCollections(images), [images]);
	const albumCollections = useMemo(() => buildAutomaticAlbumCollections(images), [images]);

	const selectedFolderId = activeSection === 'folders' ? (searchParams.get(selectionSearchParamMap.folders) ?? '') : '';
	const selectedAlbumId = activeSection === 'albums' ? (searchParams.get(selectionSearchParamMap.albums) ?? '') : '';

	const selectedFolder = folderCollections.find((collection) => collection.id === selectedFolderId) ?? null;
	const selectedAlbum = albumCollections.find((collection) => collection.id === selectedAlbumId) ?? null;

	const baseImages = useMemo(() => {
		if (activeSection === 'folders') {
			return selectedFolder?.images ?? [];
		}

		if (activeSection === 'albums') {
			return selectedAlbum?.images ?? [];
		}

		return images.filter((item) => matchesImageSection(activeSection, item, imageDates.get(item.id) ?? null));
	}, [activeSection, imageDates, images, selectedAlbum, selectedFolder]);

	const filteredImages = useMemo(
		() => baseImages.filter((item) => matchesImageSearch(item, search)),
		[baseImages, search],
	);

	const groupedImages = useMemo(() => {
		const grouped = new Map<string, { label: string; items: typeof filteredImages }>();
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
				grouped.set(key, { label, items: [] });
			}
			grouped.get(key)?.items.push(item);
		}
		return Array.from(grouped.values());
	}, [filteredImages, imageDates, imageGroupBy, monthFormatter, t]);

	const filteredFolderCards = useMemo(() => {
		const searchValue = search.trim().toLowerCase();
		return folderCollections
			.filter((collection) => {
				if (!searchValue) {
					return true;
				}

				return `${getCollectionTitleFromPath(collection.id)} ${collection.id}`.toLowerCase().includes(searchValue);
			})
			.map((collection) => ({
				id: collection.id,
				title: getCollectionTitleFromPath(collection.id),
				description: collection.id,
				imageCount: collection.images.length,
				coverImageId: collection.cover?.id,
			}));
	}, [folderCollections, search]);

	const filteredAlbumCards = useMemo(() => {
		const searchValue = search.trim().toLowerCase();
		return albumCollections
			.filter((collection) => {
				const translation = automaticAlbumTranslationKeys[collection.id as AutomaticAlbumKey];
				const title = t(translation.titleKey);
				const description = t(translation.descriptionKey);
				if (!searchValue) {
					return true;
				}

				return `${title} ${description}`.toLowerCase().includes(searchValue);
			})
			.map((collection) => {
				const translation = automaticAlbumTranslationKeys[collection.id as AutomaticAlbumKey];
				return {
					id: collection.id,
					title: t(translation.titleKey),
					description: t(translation.descriptionKey),
					imageCount: collection.images.length,
					coverImageId: collection.cover?.id,
				};
			});
	}, [albumCollections, search, t]);

	const viewMode =
		activeSection === 'folders' && !selectedFolder ? 'folders' : activeSection === 'albums' && !selectedAlbum ? 'albums' : 'grid';

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

	const updateSearchParams = useCallback(
		(updates: Record<string, string | null>) => {
			const nextSearchParams = new URLSearchParams(searchParams);
			for (const [key, value] of Object.entries(updates)) {
				if (!value) {
					nextSearchParams.delete(key);
					continue;
				}

				nextSearchParams.set(key, value);
			}
			setSearchParams(nextSearchParams);
		},
		[searchParams, setSearchParams],
	);

	const requestedImageId = Number(searchParams.get('image') ?? '');
	const activeImageDate = activeImage ? (imageDates.get(activeImage.id) ?? null) : null;

	const handleOpenImage = useCallback(
		(imageId: number) => {
			openImage(imageId);
			updateSearchParams({ image: String(imageId) });
		},
		[openImage, updateSearchParams],
	);

	const handleCloseViewer = useCallback(() => {
		closeViewer();
		updateSearchParams({ image: null });
	}, [closeViewer, updateSearchParams]);

	const handleSelectFolder = useCallback(
		(folderId: string | null) => {
			updateSearchParams({
				[selectionSearchParamMap.folders]: folderId,
				image: null,
			});
		},
		[updateSearchParams],
	);

	const handleSelectAlbum = useCallback(
		(albumId: string | null) => {
			updateSearchParams({
				[selectionSearchParamMap.albums]: albumId,
				image: null,
			});
		},
		[updateSearchParams],
	);

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
		enabled: viewMode === 'grid' && hasNextPage && !isFetchingNextPage && filteredImages.length > 0,
		rootMargin: '500px',
		onIntersect: handleLoadMore,
	});

	const activeSelection = activeSection === 'folders' ? selectedFolder : selectedAlbum;
	const activeSelectionTitle =
		activeSection === 'folders'
			? selectedFolder
				? getCollectionTitleFromPath(selectedFolder.id)
				: ''
			: selectedAlbum
				? t(automaticAlbumTranslationKeys[selectedAlbum.id as AutomaticAlbumKey].titleKey)
				: '';
	const activeSelectionDescription =
		activeSection === 'folders'
			? selectedFolder?.id ?? ''
			: selectedAlbum
				? t(automaticAlbumTranslationKeys[selectedAlbum.id as AutomaticAlbumKey].descriptionKey)
				: '';

	const titleBySection = {
		library: t('IMAGES_SECTION_LIBRARY'),
		recent: t('IMAGES_SECTION_RECENT'),
		captures: t('IMAGES_SECTION_CAPTURES'),
		photos: t('IMAGES_SECTION_PHOTOS'),
		folders: activeSelection ? activeSelectionTitle : t('IMAGES_SECTION_FOLDERS'),
		albums: activeSelection ? activeSelectionTitle : t('IMAGES_SECTION_ALBUMS'),
	} as const;

	const summary =
		viewMode === 'folders'
			? t('IMAGES_FOLDERS_SUMMARY', { filtered: String(filteredFolderCards.length), total: String(folderCollections.length) })
			: viewMode === 'albums'
				? t('IMAGES_ALBUMS_SUMMARY', { filtered: String(filteredAlbumCards.length), total: String(albumCollections.length) })
				: t('IMAGES_COUNT_SUMMARY', { filtered: String(filteredImages.length), total: String(baseImages.length) });

	const emptyState =
		viewMode === 'folders'
			? {
					title: t('IMAGES_FOLDERS_EMPTY_TITLE'),
					description: t('IMAGES_FOLDERS_EMPTY_DESC'),
				}
			: viewMode === 'albums'
				? {
						title: t('IMAGES_ALBUMS_EMPTY_TITLE'),
						description: t('IMAGES_ALBUMS_EMPTY_DESC'),
					}
				: {
						title: t('IMAGES_EMPTY_TITLE'),
						description: t('IMAGES_EMPTY_DESC'),
					};

	return {
		activeImage,
		activeImageDate,
		activeIndex,
		activeSection,
		activeSelection,
		activeSelectionDescription,
		activeSelectionTitle,
		baseImages,
		dateFormatter,
		emptyState,
		filteredAlbumCards,
		filteredFolderCards,
		filteredImages,
		goNext,
		goPrevious,
		groupByLabels,
		groupedImages,
		handleCloseViewer,
		handleOpenImage,
		handleSelectAlbum,
		handleSelectFolder,
		hasNextPage: Boolean(hasNextPage),
		imageGroupBy,
		increaseZoom,
		isFetchingNextPage,
		loadMoreRef,
		resetZoom,
		search,
		setImageGroupBy,
		setSearch,
		setShowDetails,
		showDetails,
		status,
		summary,
		title: titleBySection[activeSection],
		viewMode,
		zoom,
		decreaseZoom,
		lastVisibleImageId,
		selectedAlbum,
		selectedFolder,
	};
};
