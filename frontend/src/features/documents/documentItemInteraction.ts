import type { MouseEvent } from 'react';

export type DocumentItemInteraction = {
    isSelected: boolean;
    isSelectionActive: boolean;
    onToggleSelection: () => void;
    onToggleStar: () => void;
    onOpenMenu: (event: MouseEvent<HTMLElement>) => void;
    onContextMenu: (event: MouseEvent<HTMLElement>) => void;
};
