import { act, fireEvent, renderHook } from '@testing-library/react';
import { buildImageLibraryItem } from './imageLibraryTestFixtures';
import { useImageSelection } from './useImageSelection';

const images = [1, 2, 3, 4, 5].map((fileId) => buildImageLibraryItem({ file_id: fileId }));
const marchImages = images.slice(0, 3);

describe('useImageSelection', () => {
    it('starts empty without any backend', () => {
        const { result } = renderHook(() => useImageSelection('scope'));

        expect(result.current.selectedItems).toEqual([]);
        expect(result.current.hasSelection).toBe(false);
        expect(result.current.readMonthSelectionState(marchImages)).toBe('none');
    });

    it('toggles an image keyed by file_id', () => {
        const { result } = renderHook(() => useImageSelection('scope'));

        act(() => result.current.toggle(images[1]!));
        expect(result.current.isSelected(2)).toBe(true);

        act(() => result.current.toggle(images[1]!));
        expect(result.current.hasSelection).toBe(false);
    });

    it('selects the range between the anchor and the target within the loaded images', () => {
        const { result } = renderHook(() => useImageSelection('scope'));

        act(() => result.current.toggle(images[0]!));
        act(() => result.current.selectRange(images[3]!, images));

        expect(result.current.selectedItems.map((image) => image.file_id)).toEqual([1, 2, 3, 4]);
    });

    it('reports the month state and toggles the whole month', () => {
        const { result } = renderHook(() => useImageSelection('scope'));

        act(() => result.current.toggle(marchImages[0]!));
        expect(result.current.readMonthSelectionState(marchImages)).toBe('partial');

        act(() => result.current.toggleMonth(marchImages));
        expect(result.current.readMonthSelectionState(marchImages)).toBe('all');
        expect(result.current.selectedCount).toBe(3);

        act(() => result.current.toggleMonth(marchImages));
        expect(result.current.readMonthSelectionState(marchImages)).toBe('none');
        expect(result.current.hasSelection).toBe(false);
    });

    it('clears when the scope (filters, ordering, folder) changes', () => {
        const { result, rerender } = renderHook(({ scopeKey }) => useImageSelection(scopeKey), {
            initialProps: { scopeKey: 'all' },
        });

        act(() => result.current.selectAll(images));
        expect(result.current.selectedCount).toBe(5);

        rerender({ scopeKey: 'favorites' });
        expect(result.current.selectedCount).toBe(0);
    });

    it('clears on Escape and ignores other keys', () => {
        const { result } = renderHook(() => useImageSelection('scope'));
        act(() => result.current.selectAll(images));

        fireEvent.keyDown(document, { key: 'a' });
        expect(result.current.selectedCount).toBe(5);

        fireEvent.keyDown(document, { key: 'Escape' });
        expect(result.current.selectedCount).toBe(0);
    });
});
