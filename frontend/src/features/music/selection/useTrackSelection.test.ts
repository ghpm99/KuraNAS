import { act, renderHook } from '@testing-library/react';
import { MAX_TRACKS_PER_BATCH_EDIT, useTrackSelection } from './useTrackSelection';
import { buildTrack, buildTracks } from './trackSelectionTestFactory';

describe('useTrackSelection', () => {
    it('starts outside selection mode with nothing selected', () => {
        const { result } = renderHook(() => useTrackSelection('list'));

        expect(result.current.isSelecting).toBe(false);
        expect(result.current.selectedCount).toBe(0);
        expect(result.current.canEditSelection).toBe(false);
        expect(result.current.resolveRowSelectionMode(buildTrack(1), [])).toBeUndefined();
    });

    it('toggles a track in and out of the selection', () => {
        const { result } = renderHook(() => useTrackSelection('list'));
        act(() => result.current.startSelecting());

        act(() => result.current.toggle(buildTrack(3)));
        expect(result.current.selectedFileIds).toEqual([3]);
        expect(result.current.isSelected(3)).toBe(true);
        expect(result.current.resolveRowSelectionMode(buildTrack(3), [])).toEqual(
            expect.objectContaining({ isSelected: true })
        );

        act(() => result.current.toggle(buildTrack(3)));
        expect(result.current.selectedFileIds).toEqual([]);
    });

    it('selects every loaded track and clears them while staying in the mode', () => {
        const { result } = renderHook(() => useTrackSelection('list'));
        act(() => result.current.startSelecting());

        act(() => result.current.selectAllLoaded(buildTracks(3)));
        expect(result.current.selectedFileIds).toEqual([1, 2, 3]);
        expect(result.current.canEditSelection).toBe(true);

        act(() => result.current.clear());
        expect(result.current.selectedCount).toBe(0);
        expect(result.current.isSelecting).toBe(true);
    });

    it('leaves the mode and drops the selection when selecting stops', () => {
        const { result } = renderHook(() => useTrackSelection('list'));
        act(() => result.current.startSelecting());
        act(() => result.current.toggle(buildTrack(1)));

        act(() => result.current.stopSelecting());

        expect(result.current.isSelecting).toBe(false);
        expect(result.current.selectedCount).toBe(0);
    });

    it('allows editing exactly the limit and refuses one track more', () => {
        const { result } = renderHook(() => useTrackSelection('list'));
        act(() => result.current.startSelecting());

        act(() => result.current.selectAllLoaded(buildTracks(MAX_TRACKS_PER_BATCH_EDIT)));
        expect(result.current.isOverBatchLimit).toBe(false);
        expect(result.current.canEditSelection).toBe(true);

        act(() => result.current.selectAllLoaded(buildTracks(MAX_TRACKS_PER_BATCH_EDIT + 1)));
        expect(result.current.isOverBatchLimit).toBe(true);
        expect(result.current.canEditSelection).toBe(false);
    });

    it('discards mode and selection when the list scope changes', () => {
        const { result, rerender } = renderHook(({ scopeKey }) => useTrackSelection(scopeKey), {
            initialProps: { scopeKey: 'album:a' },
        });
        act(() => result.current.startSelecting());
        act(() => result.current.toggle(buildTrack(1)));

        rerender({ scopeKey: 'album:b' });

        expect(result.current.isSelecting).toBe(false);
        expect(result.current.selectedCount).toBe(0);
    });

    describe('range selection', () => {
        const displayedTracks = buildTracks(6);
        const clickRow = (
            selection: ReturnType<typeof useTrackSelection>,
            trackId: number,
            isRangeClick: boolean
        ) => {
            const track = buildTrack(trackId);
            selection.resolveRowSelectionMode(track, displayedTracks)?.onSelect(track, {
                isRangeClick,
            });
        };

        it('selects the tracks between the anchor and the clicked row, downwards', () => {
            const { result } = renderHook(() => useTrackSelection('list'));
            act(() => result.current.startSelecting());

            act(() => clickRow(result.current, 2, false));
            act(() => clickRow(result.current, 5, true));

            expect(result.current.selectedFileIds.sort()).toEqual([2, 3, 4, 5]);
        });

        it('selects the range upwards when the clicked row is above the anchor', () => {
            const { result } = renderHook(() => useTrackSelection('list'));
            act(() => result.current.startSelecting());

            act(() => clickRow(result.current, 5, false));
            act(() => clickRow(result.current, 3, true));

            expect(result.current.selectedFileIds.sort()).toEqual([3, 4, 5]);
        });

        it('toggles a single row when the click has no shift', () => {
            const { result } = renderHook(() => useTrackSelection('list'));
            act(() => result.current.startSelecting());

            act(() => clickRow(result.current, 4, false));
            act(() => clickRow(result.current, 4, false));

            expect(result.current.selectedCount).toBe(0);
        });
    });
});
