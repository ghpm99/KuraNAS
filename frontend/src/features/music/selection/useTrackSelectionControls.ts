import { useEffect, useState } from 'react';
import type { IMusicData } from '@/types/music';
import type { TrackSelection } from './useTrackSelection';

const isTypingTarget = (target: EventTarget | null) =>
    target instanceof HTMLElement &&
    (target.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName));

export const useTrackSelectionControls = (
    selection: TrackSelection,
    loadedTracks: IMusicData[]
) => {
    const { isSelecting, selectedFileIds, stopSelecting, selectAllLoaded } = selection;
    const [editedFileIds, setEditedFileIds] = useState<number[] | null>(null);
    const isEditDialogOpen = editedFileIds !== null;

    useEffect(() => {
        if (!isSelecting || isEditDialogOpen) return undefined;
        const handleSelectionShortcut = (event: KeyboardEvent) => {
            if (event.key === 'Escape') {
                stopSelecting();
                return;
            }
            const isSelectAllShortcut =
                (event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'a';
            if (!isSelectAllShortcut || isTypingTarget(event.target)) return;
            event.preventDefault();
            selectAllLoaded(loadedTracks);
        };
        document.addEventListener('keydown', handleSelectionShortcut);
        return () => document.removeEventListener('keydown', handleSelectionShortcut);
    }, [isSelecting, isEditDialogOpen, stopSelecting, selectAllLoaded, loadedTracks]);

    return {
        editedFileIds,
        openEditDialog: () => setEditedFileIds(selectedFileIds),
        closeEditDialog: () => setEditedFileIds(null),
    };
};

export default useTrackSelectionControls;
