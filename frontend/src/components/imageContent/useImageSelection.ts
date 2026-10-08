import { useCallback, useEffect } from 'react';
import { useItemSelection, type ItemSelection } from '@/shared/selection/useItemSelection';
import type { ImageLibraryItem } from '@/types/imageLibrary';
import { readImageLibraryItemId } from './imageLibraryItemMapping';

export type MonthSelectionState = 'none' | 'partial' | 'all';

export type ImageSelection = ItemSelection<ImageLibraryItem> & {
    readMonthSelectionState: (monthImages: ImageLibraryItem[]) => MonthSelectionState;
    toggleMonth: (monthImages: ImageLibraryItem[]) => void;
};

export const useImageSelection = (scopeKey: string): ImageSelection => {
    const selection = useItemSelection<ImageLibraryItem>(scopeKey, readImageLibraryItemId);
    const { hasSelection, clear, isSelected, select, deselect } = selection;

    useEffect(() => {
        if (!hasSelection) {
            return undefined;
        }
        const clearOnEscape = (event: KeyboardEvent) => {
            if (event.key === 'Escape') {
                clear();
            }
        };
        document.addEventListener('keydown', clearOnEscape);
        return () => document.removeEventListener('keydown', clearOnEscape);
    }, [hasSelection, clear]);

    const readMonthSelectionState = useCallback(
        (monthImages: ImageLibraryItem[]): MonthSelectionState => {
            const selectedInMonth = monthImages.filter((image) => isSelected(image.file_id)).length;
            if (selectedInMonth === 0) {
                return 'none';
            }
            return selectedInMonth === monthImages.length ? 'all' : 'partial';
        },
        [isSelected]
    );

    const toggleMonth = useCallback(
        (monthImages: ImageLibraryItem[]) => {
            if (readMonthSelectionState(monthImages) === 'all') {
                deselect(monthImages);
                return;
            }
            select(monthImages);
        },
        [readMonthSelectionState, select, deselect]
    );

    return { ...selection, readMonthSelectionState, toggleMonth };
};
