import { ListItemIcon, ListItemText, Menu, MenuItem } from '@mui/material';
import type { ReactNode } from 'react';
import type { MenuPosition } from './useMenuPosition';

export type MusicContextMenuAction = {
    key: string;
    label: string;
    icon: ReactNode;
    onSelect: () => void;
};

type MusicContextMenuProps = {
    position: MenuPosition | null;
    actions: MusicContextMenuAction[];
    onClose: () => void;
};

export default function MusicContextMenu({
    position,
    actions = [],
    onClose,
}: MusicContextMenuProps) {
    return (
        <Menu
            open={position !== null}
            onClose={onClose}
            anchorReference="anchorPosition"
            anchorPosition={position ?? undefined}
            onClick={(event) => event.stopPropagation()}
        >
            {actions.map((action) => (
                <MenuItem
                    key={action.key}
                    onClick={() => {
                        onClose();
                        action.onSelect();
                    }}
                >
                    <ListItemIcon>{action.icon}</ListItemIcon>
                    <ListItemText primary={action.label} />
                </MenuItem>
            ))}
        </Menu>
    );
}
