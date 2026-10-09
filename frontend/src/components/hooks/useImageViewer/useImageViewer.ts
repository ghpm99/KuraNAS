import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { isTypingInTextField } from './isTextEditingTarget';

type IdentifiableImage = { id: number };

const defaultSlideshowIntervalInMs = 3500;
const minimumZoom = 0.5;
const maximumZoom = 5;
const zoomStep = 0.2;
const fullTurnInDegrees = 360;
const rotationStepInDegrees = 90;

const readIdentifiableImageId = (image: IdentifiableImage) => image.id;

const clampZoom = (zoomLevel: number) =>
    Math.min(maximumZoom, Math.max(minimumZoom, Number(zoomLevel.toFixed(2))));

export type ImageViewerPaging<T> = {
    hasNext: boolean;
    hasPrevious: boolean;
    isLoading: boolean;
    loadNext: () => Promise<T[] | undefined>;
    loadPrevious: () => Promise<T[] | undefined>;
};

const noPaging: ImageViewerPaging<never> = {
    hasNext: false,
    hasPrevious: false,
    isLoading: false,
    loadNext: () => Promise.resolve(undefined),
    loadPrevious: () => Promise.resolve(undefined),
};

type ViewTransform = { zoom: number; panX: number; panY: number };

const untransformedView: ViewTransform = { zoom: 1, panX: 0, panY: 0 };

const withZoom = (view: ViewTransform, zoomLevel: number): ViewTransform =>
    zoomLevel <= 1 ? { zoom: zoomLevel, panX: 0, panY: 0 } : { ...view, zoom: zoomLevel };

type ImageRotation = { imageId: number | null; degrees: number };

const unrotatedImage: ImageRotation = { imageId: null, degrees: 0 };

export function useImageViewer<T>(
    images: T[],
    slideshowIntervalInMs = defaultSlideshowIntervalInMs,
    getImageId: (image: T) => number = readIdentifiableImageId as (image: T) => number,
    paging: ImageViewerPaging<T> = noPaging
) {
    const [viewerImageId, setViewerImageId] = useState<number | null>(null);
    const [{ zoom, panX, panY }, setViewTransform] = useState<ViewTransform>(untransformedView);
    const [showDetails, setShowDetails] = useState(true);
    const [showFilmstrip, setShowFilmstrip] = useState(true);
    const [isSlideshowPlaying, setIsSlideshowPlaying] = useState(false);
    const [imageRotation, setImageRotation] = useState<ImageRotation>(unrotatedImage);
    const isLoadingMoreRef = useRef(false);

    const activeIndex = useMemo(
        () => images.findIndex((image) => getImageId(image) === viewerImageId),
        [images, viewerImageId, getImageId]
    );
    const activeImage = activeIndex >= 0 ? images[activeIndex] : null;
    const rotation = imageRotation.imageId === viewerImageId ? imageRotation.degrees : 0;
    const canGoNext = activeIndex >= 0 && (activeIndex < images.length - 1 || paging.hasNext);
    const canGoPrevious = activeIndex > 0 || (activeIndex === 0 && paging.hasPrevious);

    const showImage = useCallback((id: number) => {
        setViewerImageId(id);
        setViewTransform(untransformedView);
        setImageRotation(unrotatedImage);
    }, []);

    const openImage = showImage;

    const closeViewer = useCallback(() => {
        setViewerImageId(null);
        setViewTransform(untransformedView);
        setImageRotation(unrotatedImage);
        setIsSlideshowPlaying(false);
    }, []);

    const moveAfterLoading = useCallback(
        async (
            loadMore: () => Promise<T[] | undefined>,
            offsetFromActive: 1 | -1,
            activeImageId: number
        ) => {
            if (isLoadingMoreRef.current || paging.isLoading) {
                return;
            }
            isLoadingMoreRef.current = true;
            try {
                const loadedImages = await loadMore();
                const loadedIndex =
                    loadedImages?.findIndex((image) => getImageId(image) === activeImageId) ?? -1;
                const targetImage = loadedImages?.[loadedIndex + offsetFromActive];
                if (loadedIndex < 0 || !targetImage) {
                    return;
                }
                setViewerImageId((currentImageId) =>
                    currentImageId === activeImageId ? getImageId(targetImage) : currentImageId
                );
                setViewTransform(untransformedView);
                setImageRotation(unrotatedImage);
            } finally {
                isLoadingMoreRef.current = false;
            }
        },
        [getImageId, paging.isLoading]
    );

    const goNext = useCallback(() => {
        if (!activeImage || activeIndex < 0) return;
        const nextImage = images[activeIndex + 1];
        if (nextImage) {
            showImage(getImageId(nextImage));
            return;
        }
        if (paging.hasNext) {
            void moveAfterLoading(paging.loadNext, 1, getImageId(activeImage));
        }
    }, [activeImage, activeIndex, images, getImageId, paging, showImage, moveAfterLoading]);

    const goPrevious = useCallback(() => {
        if (!activeImage || activeIndex < 0) return;
        const previousImage = images[activeIndex - 1];
        if (previousImage) {
            showImage(getImageId(previousImage));
            return;
        }
        if (paging.hasPrevious) {
            void moveAfterLoading(paging.loadPrevious, -1, getImageId(activeImage));
        }
    }, [activeImage, activeIndex, images, getImageId, paging, showImage, moveAfterLoading]);

    const increaseZoom = useCallback(() => {
        setViewTransform((view) => withZoom(view, clampZoom(view.zoom + zoomStep)));
    }, []);

    const decreaseZoom = useCallback(() => {
        setViewTransform((view) => withZoom(view, clampZoom(view.zoom - zoomStep)));
    }, []);

    const resetZoom = useCallback(() => {
        setViewTransform(untransformedView);
    }, []);

    const setZoomLevel = useCallback((zoomLevel: number) => {
        setViewTransform((view) => withZoom(view, clampZoom(zoomLevel)));
    }, []);

    const setPan = useCallback((nextPanX: number, nextPanY: number) => {
        setViewTransform((view) =>
            view.zoom > 1 ? { ...view, panX: nextPanX, panY: nextPanY } : view
        );
    }, []);

    const rotateClockwise = useCallback(() => {
        setImageRotation((current) => {
            const currentDegrees = current.imageId === viewerImageId ? current.degrees : 0;
            return {
                imageId: viewerImageId,
                degrees: (currentDegrees + rotationStepInDegrees) % fullTurnInDegrees,
            };
        });
    }, [viewerImageId]);

    const toggleSlideshow = useCallback(() => {
        if (images.length <= 1) {
            return;
        }

        setIsSlideshowPlaying((value) => !value);
    }, [images.length]);

    useEffect(() => {
        if (!activeImage) return;

        const onKeyDown = (event: KeyboardEvent) => {
            if (isTypingInTextField(event) || event.ctrlKey || event.metaKey || event.altKey) {
                return;
            }
            if (event.key === 'Escape') closeViewer();
            if (event.key === 'ArrowRight') goNext();
            if (event.key === 'ArrowLeft') goPrevious();
            if (event.key === '+' || event.key === '=') increaseZoom();
            if (event.key === '-') decreaseZoom();
            if (event.key === '0') resetZoom();
            if (event.key.toLowerCase() === 'i') setShowDetails((value) => !value);
            if (event.key.toLowerCase() === 'f') setShowFilmstrip((value) => !value);
            if (event.key.toLowerCase() === 's') toggleSlideshow();
            if (event.key.toLowerCase() === 'r') rotateClockwise();
        };

        window.addEventListener('keydown', onKeyDown);
        return () => window.removeEventListener('keydown', onKeyDown);
    }, [
        activeImage,
        closeViewer,
        decreaseZoom,
        goNext,
        goPrevious,
        increaseZoom,
        resetZoom,
        rotateClockwise,
        toggleSlideshow,
    ]);

    useEffect(() => {
        if (!activeImage || !isSlideshowPlaying || images.length <= 1) {
            return;
        }

        const intervalId = window.setInterval(() => {
            if (!canGoNext) {
                setIsSlideshowPlaying(false);
                return;
            }
            goNext();
        }, slideshowIntervalInMs);

        return () => window.clearInterval(intervalId);
    }, [activeImage, canGoNext, goNext, images.length, isSlideshowPlaying, slideshowIntervalInMs]);

    return {
        viewerImageId,
        activeImage,
        activeIndex,
        zoom,
        pan: { x: panX, y: panY },
        rotation,
        showDetails,
        showFilmstrip,
        isSlideshowPlaying,
        canGoNext,
        canGoPrevious,
        setShowDetails,
        setShowFilmstrip,
        openImage,
        closeViewer,
        goNext,
        goPrevious,
        increaseZoom,
        decreaseZoom,
        resetZoom,
        setZoomLevel,
        setPan,
        rotateClockwise,
        toggleSlideshow,
    };
}
