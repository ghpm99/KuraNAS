import { act, renderHook, waitFor } from '@testing-library/react';
import { getImageLibraryNeighbors } from '@/service/image';
import type { ImageLibraryFilters } from '@/types/imageLibrary';
import { buildImageLibraryItem } from './imageLibraryTestFixtures';
import { useImageViewerNeighbors } from './useImageViewerNeighbors';

const noFilters: ImageLibraryFilters = {
    nameQuery: '',
    categories: [],
    isStarredOnly: false,
    formats: [],
    camera: '',
    takenFrom: '',
    takenTo: '',
    folder: '',
};

const imageWithId = (fileId: number) =>
    buildImageLibraryItem({ file_id: fileId, name: `Photo-${fileId}.jpg` });

const imagesWithIds = (fileIds: number[]) => fileIds.map(imageWithId);

jest.mock('@/service/image', () => ({
    defaultImageNeighborsCount: 2,
    getImageLibraryNeighbors: jest.fn(),
}));

const mockedNeighbors = getImageLibraryNeighbors as jest.Mock;

describe('useImageViewerNeighbors', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('lists the newer neighbors, the pivot and the older neighbors in gallery order', async () => {
        mockedNeighbors.mockResolvedValue({
            before: imagesWithIds([12, 11]),
            after: imagesWithIds([9]),
        });
        const pivot = imageWithId(10);

        const { result } = renderHook(() => useImageViewerNeighbors({ pivot, filters: noFilters }));

        await waitFor(() =>
            expect(result.current.neighborImages.map((image) => image.file_id)).toEqual([
                12, 11, 10, 9,
            ])
        );
        expect(mockedNeighbors).toHaveBeenCalledWith(10, noFilters);
        expect(result.current.neighborsPaging.hasPrevious).toBe(true);
        expect(result.current.neighborsPaging.hasNext).toBe(false);
    });

    it('requests the neighbors of a pivot only once', async () => {
        mockedNeighbors.mockResolvedValue({ before: [], after: [] });
        const pivot = imageWithId(10);

        const { rerender } = renderHook(() =>
            useImageViewerNeighbors({ pivot, filters: noFilters })
        );
        rerender();
        rerender();

        await waitFor(() => expect(mockedNeighbors).toHaveBeenCalledTimes(1));
    });

    it('extends forwards from the last image without duplicating items', async () => {
        mockedNeighbors
            .mockResolvedValueOnce({ before: [], after: imagesWithIds([9, 8]) })
            .mockResolvedValueOnce({ before: [], after: imagesWithIds([8, 7]) });
        const pivot = imageWithId(10);
        const { result } = renderHook(() => useImageViewerNeighbors({ pivot, filters: noFilters }));
        await waitFor(() => expect(result.current.neighborsPaging.hasNext).toBe(true));

        let mergedImages: unknown[] | undefined;
        await act(async () => {
            mergedImages = await result.current.neighborsPaging.loadNext();
        });

        expect(mockedNeighbors).toHaveBeenLastCalledWith(8, noFilters);
        expect(mergedImages).toHaveLength(4);
        expect(result.current.neighborImages.map((image) => image.file_id)).toEqual([10, 9, 8, 7]);
        expect(result.current.neighborsPaging.hasNext).toBe(true);
    });

    it('extends backwards from the first image and stops when the end is reached', async () => {
        mockedNeighbors
            .mockResolvedValueOnce({ before: imagesWithIds([12, 11]), after: [] })
            .mockResolvedValueOnce({ before: imagesWithIds([13]), after: [] });
        const pivot = imageWithId(10);
        const { result } = renderHook(() => useImageViewerNeighbors({ pivot, filters: noFilters }));
        await waitFor(() => expect(result.current.neighborsPaging.hasPrevious).toBe(true));

        await act(async () => {
            await result.current.neighborsPaging.loadPrevious();
        });

        expect(mockedNeighbors).toHaveBeenLastCalledWith(12, noFilters);
        expect(result.current.neighborImages.map((image) => image.file_id)).toEqual([
            13, 12, 11, 10,
        ]);
        expect(result.current.neighborsPaging.hasPrevious).toBe(false);
    });

    it('keeps the list and reports undefined when extending fails', async () => {
        mockedNeighbors
            .mockResolvedValueOnce({
                before: imagesWithIds([12, 11]),
                after: imagesWithIds([9, 8]),
            })
            .mockRejectedValue(new Error('offline'));
        const pivot = imageWithId(10);
        const { result } = renderHook(() => useImageViewerNeighbors({ pivot, filters: noFilters }));
        await waitFor(() => expect(result.current.neighborsPaging.hasNext).toBe(true));

        let forward: unknown[] | undefined = [];
        let backward: unknown[] | undefined = [];
        await act(async () => {
            forward = await result.current.neighborsPaging.loadNext();
            backward = await result.current.neighborsPaging.loadPrevious();
        });

        expect(forward).toBeUndefined();
        expect(backward).toBeUndefined();
        expect(result.current.neighborImages).toHaveLength(5);
        expect(result.current.neighborsPaging.isLoading).toBe(false);
    });

    it('has nothing to extend before a pivot was resolved and clears on demand', async () => {
        const idle = renderHook(() => useImageViewerNeighbors({ pivot: null, filters: noFilters }));
        let nothing: unknown[] | undefined = [];
        await act(async () => {
            nothing = await idle.result.current.neighborsPaging.loadNext();
            await idle.result.current.neighborsPaging.loadPrevious();
        });
        expect(nothing).toBeUndefined();
        expect(mockedNeighbors).not.toHaveBeenCalled();

        mockedNeighbors.mockResolvedValue({ before: [], after: imagesWithIds([9]) });
        const pivot = imageWithId(10);
        const { result } = renderHook(() => useImageViewerNeighbors({ pivot, filters: noFilters }));
        await waitFor(() => expect(result.current.neighborImages).toHaveLength(2));

        act(() => {
            result.current.clearNeighbors();
        });
        expect(result.current.neighborImages).toEqual([]);
    });
});
