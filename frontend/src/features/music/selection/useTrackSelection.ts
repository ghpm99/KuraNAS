import { useCallback, useState } from 'react';
import { useItemSelection } from '@/shared/selection/useItemSelection';
import type { IMusicData } from '@/types/music';

export const MAX_TRACKS_PER_BATCH_EDIT = 500;

const readTrackId = (track: IMusicData) => track.id;

export type TrackRowSelectionMode = {
    isSelected: boolean;
    onSelect: (track: IMusicData, click: { isRangeClick: boolean }) => void;
    onToggle: (track: IMusicData) => void;
};

export type TrackSelection = {
    isSelecting: boolean;
    selectedTracks: IMusicData[];
    selectedFileIds: number[];
    selectedCount: number;
    isOverBatchLimit: boolean;
    canEditSelection: boolean;
    isSelected: (trackId: number) => boolean;
    toggle: (track: IMusicData) => void;
    selectAllLoaded: (loadedTracks: IMusicData[]) => void;
    clear: () => void;
    resolveRowSelectionMode: (
        track: IMusicData,
        displayedTracks: IMusicData[]
    ) => TrackRowSelectionMode | undefined;
    startSelecting: () => void;
    stopSelecting: () => void;
};

/**
 * Selection state of one track list. The state belongs to `scopeKey`: when the
 * key changes the mode and the selected tracks are discarded.
 */
export const useTrackSelection = (scopeKey: string): TrackSelection => {
    const { selectedItems, selectedCount, isSelected, toggle, selectRange, selectAll, clear } =
        useItemSelection<IMusicData>(scopeKey, readTrackId);
    const [selectingScopeKey, setSelectingScopeKey] = useState<string | null>(null);

    const startSelecting = useCallback(() => setSelectingScopeKey(scopeKey), [scopeKey]);

    const stopSelecting = useCallback(() => {
        setSelectingScopeKey(null);
        clear();
    }, [clear]);

    const isSelecting = selectingScopeKey === scopeKey;

    const resolveRowSelectionMode = (track: IMusicData, displayedTracks: IMusicData[]) =>
        isSelecting
            ? {
                  isSelected: isSelected(track.id),
                  onSelect: (
                      clickedTrack: IMusicData,
                      { isRangeClick }: { isRangeClick: boolean }
                  ) =>
                      isRangeClick
                          ? selectRange(clickedTrack, displayedTracks)
                          : toggle(clickedTrack),
                  onToggle: toggle,
              }
            : undefined;

    const isOverBatchLimit = selectedCount > MAX_TRACKS_PER_BATCH_EDIT;

    return {
        isSelecting,
        selectedTracks: selectedItems,
        selectedFileIds: selectedItems.map(readTrackId),
        selectedCount,
        isOverBatchLimit,
        canEditSelection: selectedCount > 0 && !isOverBatchLimit,
        isSelected,
        toggle,
        selectAllLoaded: selectAll,
        clear,
        resolveRowSelectionMode,
        startSelecting,
        stopSelecting,
    };
};

export default useTrackSelection;
