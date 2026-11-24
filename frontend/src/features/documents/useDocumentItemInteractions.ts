import { useState, type MouseEvent } from 'react';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import useFileOperations from '@/features/files/fileActions/useFileOperations';
import { useFileSelectionContext } from '@/features/files/selection/fileSelectionContext';
import type { DocumentItemMenuAnchor } from './DocumentItemMenu';
import type { DocumentItemInteraction } from './documentItemInteraction';

type OpenMenuState = { file: FileData; anchorPosition: DocumentItemMenuAnchor };

const useDocumentItemInteractions = () => {
    const { toggleFavorites } = useFileOperations();
    const { isSelected, hasSelection, toggle, selectedFiles, selectedCount } =
        useFileSelectionContext();
    const [openMenu, setOpenMenu] = useState<OpenMenuState | null>(null);

    const interactionFor = (file: FileData): DocumentItemInteraction => ({
        isSelected: isSelected(file.id),
        isSelectionActive: hasSelection,
        onToggleSelection: () => toggle(file),
        onToggleStar: () => {
            void toggleFavorites([file]);
        },
        onOpenMenu: (event: MouseEvent<HTMLElement>) => {
            event.stopPropagation();
            const buttonBounds = event.currentTarget.getBoundingClientRect();
            setOpenMenu({
                file,
                anchorPosition: { top: buttonBounds.bottom, left: buttonBounds.left },
            });
        },
        onContextMenu: (event: MouseEvent<HTMLElement>) => {
            event.preventDefault();
            setOpenMenu({
                file,
                anchorPosition: { top: event.clientY, left: event.clientX },
            });
        },
    });

    const menuTargetFiles = (): FileData[] => {
        if (!openMenu) return [];
        const isPartOfMultiSelection = isSelected(openMenu.file.id) && selectedCount > 1;
        return isPartOfMultiSelection ? selectedFiles : [openMenu.file];
    };

    return {
        interactionFor,
        menuAnchorPosition: openMenu?.anchorPosition ?? null,
        menuTargetFiles: menuTargetFiles(),
        closeMenu: () => setOpenMenu(null),
    };
};

export default useDocumentItemInteractions;
