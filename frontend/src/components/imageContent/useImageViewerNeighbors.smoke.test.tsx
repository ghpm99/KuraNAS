import { act, renderHook } from '@testing-library/react';
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

describe('useImageViewerNeighbors without any backend mock', () => {
    it('stays empty without a pivot and survives a failing backend', async () => {
        const withoutPivot = renderHook(() =>
            useImageViewerNeighbors({ pivot: null, filters: noFilters })
        );
        expect(withoutPivot.result.current.neighborImages).toEqual([]);
        expect(withoutPivot.result.current.neighborsPaging.hasNext).toBe(false);

        const pivot = buildImageLibraryItem({ file_id: 5 });
        const withPivot = renderHook(() => useImageViewerNeighbors({ pivot, filters: noFilters }));
        await act(async () => {
            await new Promise((resolve) => setTimeout(resolve, 20));
        });
        expect(withPivot.result.current.neighborImages).toEqual([]);
    });
});
