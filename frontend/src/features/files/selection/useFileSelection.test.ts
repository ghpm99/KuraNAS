import { act, renderHook } from '@testing-library/react';
import { useFileSelection } from './useFileSelection';
import { createTestFile } from './testFileFactory';

const listedFiles = [1, 2, 3, 4, 5].map((id) => createTestFile(id));

describe('useFileSelection', () => {
    it('starts empty without any backend', () => {
        const { result } = renderHook(() => useFileSelection('root'));

        expect(result.current.selectedFiles).toEqual([]);
        expect(result.current.selectedCount).toBe(0);
        expect(result.current.hasSelection).toBe(false);
        expect(result.current.isSelected(1)).toBe(false);
    });

    it('toggles a file on and off', () => {
        const { result } = renderHook(() => useFileSelection('root'));

        act(() => result.current.toggle(listedFiles[0]!));
        expect(result.current.isSelected(1)).toBe(true);
        expect(result.current.hasSelection).toBe(true);

        act(() => result.current.toggle(listedFiles[0]!));
        expect(result.current.isSelected(1)).toBe(false);
        expect(result.current.hasSelection).toBe(false);
    });

    it('selects the range between the last anchor and the target, in either direction', () => {
        const { result } = renderHook(() => useFileSelection('root'));

        act(() => result.current.toggle(listedFiles[3]!));
        act(() => result.current.selectRange(listedFiles[1]!, listedFiles));

        expect(result.current.selectedFiles.map((file) => file.id).sort()).toEqual([2, 3, 4]);
    });

    it('selects only the target when there is no anchor in the list', () => {
        const { result } = renderHook(() => useFileSelection('root'));

        act(() => result.current.selectRange(listedFiles[2]!, listedFiles));
        expect(result.current.selectedFiles.map((file) => file.id)).toEqual([3]);

        act(() => result.current.selectRange(listedFiles[2]!, listedFiles));
        expect(result.current.selectedFiles.map((file) => file.id)).toEqual([3]);

        act(() => result.current.selectRange(listedFiles[0]!, listedFiles));
        expect(result.current.selectedFiles.map((file) => file.id).sort()).toEqual([1, 2, 3]);
    });

    it('selects all, deselects some and clears', () => {
        const { result } = renderHook(() => useFileSelection('root'));

        act(() => result.current.selectAll(listedFiles));
        expect(result.current.selectedCount).toBe(5);

        act(() => result.current.deselect([listedFiles[0]!, listedFiles[1]!]));
        expect(result.current.selectedCount).toBe(3);

        act(() => result.current.clear());
        expect(result.current.selectedCount).toBe(0);
    });

    it('clears the selection when the folder scope changes', () => {
        const { result, rerender } = renderHook(({ scopeKey }) => useFileSelection(scopeKey), {
            initialProps: { scopeKey: 'folder-1:all' },
        });

        act(() => result.current.selectAll(listedFiles));
        expect(result.current.selectedCount).toBe(5);

        rerender({ scopeKey: 'folder-2:all' });
        expect(result.current.selectedCount).toBe(0);

        act(() => result.current.toggle(listedFiles[0]!));
        expect(result.current.selectedFiles.map((file) => file.id)).toEqual([1]);
    });
});
