import { useCallback, useEffect, useMemo, useState } from 'react';

type IdentifiableImage = { id: number };

const defaultSlideshowIntervalInMs = 3500;

export function useImageViewer<T extends IdentifiableImage>(images: T[], slideshowIntervalInMs = defaultSlideshowIntervalInMs) {
	const [viewerImageId, setViewerImageId] = useState<number | null>(null);
	const [zoom, setZoom] = useState(1);
	const [showDetails, setShowDetails] = useState(true);
	const [showFilmstrip, setShowFilmstrip] = useState(true);
	const [isSlideshowPlaying, setIsSlideshowPlaying] = useState(false);

	const activeIndex = useMemo(() => images.findIndex((image) => image.id === viewerImageId), [images, viewerImageId]);
	const activeImage = activeIndex >= 0 ? images[activeIndex] : null;

	const openImage = useCallback((id: number) => {
		setViewerImageId(id);
		setZoom(1);
	}, []);

	const closeViewer = useCallback(() => {
		setViewerImageId(null);
		setZoom(1);
		setIsSlideshowPlaying(false);
	}, []);

	const goNext = useCallback(() => {
		if (images.length === 0 || activeIndex < 0) return;
		const nextImage = images[(activeIndex + 1) % images.length];
		if (!nextImage) return;
		setViewerImageId(nextImage.id);
		setZoom(1);
	}, [images, activeIndex]);

	const goPrevious = useCallback(() => {
		if (images.length === 0 || activeIndex < 0) return;
		const previous = activeIndex === 0 ? images.length - 1 : activeIndex - 1;
		const previousImage = images[previous];
		if (!previousImage) return;
		setViewerImageId(previousImage.id);
		setZoom(1);
	}, [images, activeIndex]);

	const increaseZoom = useCallback(() => {
		setZoom((value) => Math.min(5, Number((value + 0.2).toFixed(2))));
	}, []);

	const decreaseZoom = useCallback(() => {
		setZoom((value) => Math.max(0.5, Number((value - 0.2).toFixed(2))));
	}, []);

	const resetZoom = useCallback(() => {
		setZoom(1);
	}, []);

	const toggleSlideshow = useCallback(() => {
		if (images.length <= 1) {
			return;
		}

		setIsSlideshowPlaying((value) => !value);
	}, [images.length]);

	useEffect(() => {
		if (!activeImage) return;

		const onKeyDown = (event: KeyboardEvent) => {
			if (event.key === 'Escape') closeViewer();
			if (event.key === 'ArrowRight') goNext();
			if (event.key === 'ArrowLeft') goPrevious();
			if (event.key === '+' || event.key === '=') increaseZoom();
			if (event.key === '-') decreaseZoom();
			if (event.key === '0') resetZoom();
			if (event.key.toLowerCase() === 'i') setShowDetails((value) => !value);
			if (event.key.toLowerCase() === 'f') setShowFilmstrip((value) => !value);
			if (event.key.toLowerCase() === 's') toggleSlideshow();
		};

		window.addEventListener('keydown', onKeyDown);
		return () => window.removeEventListener('keydown', onKeyDown);
	}, [activeImage, closeViewer, decreaseZoom, goNext, goPrevious, increaseZoom, resetZoom, toggleSlideshow]);

	useEffect(() => {
		if (!activeImage || !isSlideshowPlaying || images.length <= 1) {
			return;
		}

		const intervalId = window.setInterval(() => {
			goNext();
		}, slideshowIntervalInMs);

		return () => window.clearInterval(intervalId);
	}, [activeImage, goNext, images.length, isSlideshowPlaying, slideshowIntervalInMs]);

	return {
		viewerImageId,
		activeImage,
		activeIndex,
		zoom,
		showDetails,
		showFilmstrip,
		isSlideshowPlaying,
		setShowDetails,
		setShowFilmstrip,
		openImage,
		closeViewer,
		goNext,
		goPrevious,
		increaseZoom,
		decreaseZoom,
		resetZoom,
		toggleSlideshow,
	};
}
