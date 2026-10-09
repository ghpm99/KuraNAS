import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { ImageViewerPaging } from '@/components/hooks/useImageViewer/useImageViewer';
import { defaultImageNeighborsCount, getImageLibraryNeighbors } from '@/service/image';
import type { ImageLibraryFilters, ImageLibraryItem } from '@/types/imageLibrary';

type NeighborsState = {
    images: ImageLibraryItem[];
    hasMoreBefore: boolean;
    hasMoreAfter: boolean;
    isLoading: boolean;
};

const emptyNeighborsState: NeighborsState = {
    images: [],
    hasMoreBefore: false,
    hasMoreAfter: false,
    isLoading: false,
};

const withoutDuplicates = (
    images: ImageLibraryItem[],
    alreadyListed: ImageLibraryItem[]
): ImageLibraryItem[] => {
    const listedIds = new Set(alreadyListed.map((image) => image.file_id));
    return images.filter((image) => !listedIds.has(image.file_id));
};

type UseImageViewerNeighborsParams = {
    pivot: ImageLibraryItem | null;
    filters: ImageLibraryFilters;
};

export const useImageViewerNeighbors = ({ pivot, filters }: UseImageViewerNeighborsParams) => {
    const [state, setState] = useState<NeighborsState>(emptyNeighborsState);
    const requestedPivotId = useRef<number | null>(null);

    useEffect(() => {
        if (!pivot || requestedPivotId.current === pivot.file_id) {
            return;
        }
        requestedPivotId.current = pivot.file_id;
        getImageLibraryNeighbors(pivot.file_id, filters)
            .then((neighbors) => {
                setState({
                    images: [...neighbors.before, pivot, ...neighbors.after],
                    hasMoreBefore: neighbors.before.length >= defaultImageNeighborsCount,
                    hasMoreAfter: neighbors.after.length >= defaultImageNeighborsCount,
                    isLoading: false,
                });
            })
            .catch(() => undefined);
    }, [pivot, filters]);

    const loadNext = useCallback(async () => {
        const lastImage = state.images[state.images.length - 1];
        if (!lastImage) {
            return undefined;
        }
        setState((current) => ({ ...current, isLoading: true }));
        try {
            const neighbors = await getImageLibraryNeighbors(lastImage.file_id, filters);
            const mergedImages = [
                ...state.images,
                ...withoutDuplicates(neighbors.after, state.images),
            ];
            setState((current) => ({
                ...current,
                images: mergedImages,
                hasMoreAfter: neighbors.after.length >= defaultImageNeighborsCount,
                isLoading: false,
            }));
            return mergedImages;
        } catch {
            setState((current) => ({ ...current, isLoading: false }));
            return undefined;
        }
    }, [state.images, filters]);

    const loadPrevious = useCallback(async () => {
        const firstImage = state.images[0];
        if (!firstImage) {
            return undefined;
        }
        setState((current) => ({ ...current, isLoading: true }));
        try {
            const neighbors = await getImageLibraryNeighbors(firstImage.file_id, filters);
            const mergedImages = [
                ...withoutDuplicates(neighbors.before, state.images),
                ...state.images,
            ];
            setState((current) => ({
                ...current,
                images: mergedImages,
                hasMoreBefore: neighbors.before.length >= defaultImageNeighborsCount,
                isLoading: false,
            }));
            return mergedImages;
        } catch {
            setState((current) => ({ ...current, isLoading: false }));
            return undefined;
        }
    }, [state.images, filters]);

    const clear = useCallback(() => {
        requestedPivotId.current = null;
        setState(emptyNeighborsState);
    }, []);

    const paging = useMemo<ImageViewerPaging<ImageLibraryItem>>(
        () => ({
            hasNext: state.hasMoreAfter,
            hasPrevious: state.hasMoreBefore,
            isLoading: state.isLoading,
            loadNext,
            loadPrevious,
        }),
        [state.hasMoreAfter, state.hasMoreBefore, state.isLoading, loadNext, loadPrevious]
    );

    return { neighborImages: state.images, neighborsPaging: paging, clearNeighbors: clear };
};
