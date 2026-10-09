import { useState, type MouseEvent, type ReactNode } from 'react';
import { IconButton, ListItemIcon, ListItemText, Menu, MenuItem, Tooltip } from '@mui/material';
import { EllipsisVertical } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';

export type ActionBarMenuEntry = {
    key: string;
    label: string;
    icon: ReactNode;
    onSelect: () => void;
    isDestructive?: boolean;
};

type ActionBarMoreMenuProps = {
    entries: ActionBarMenuEntry[];
};

const ActionBarMoreMenu = ({ entries }: ActionBarMoreMenuProps) => {
    const { t } = useI18n();
    const [anchorElement, setAnchorElement] = useState<HTMLElement | null>(null);
    const label = t('FILES_MORE_ACTIONS');

    if (entries.length === 0) return null;

    const closeMenu = () => setAnchorElement(null);

    const selectEntry = (entry: ActionBarMenuEntry) => {
        closeMenu();
        entry.onSelect();
    };

    return (
        <>
            <Tooltip title={label}>
                <IconButton
                    size="small"
                    aria-label={label}
                    aria-haspopup="menu"
                    aria-expanded={anchorElement ? true : undefined}
                    onClick={(event: MouseEvent<HTMLElement>) =>
                        setAnchorElement(event.currentTarget)
                    }
                >
                    <EllipsisVertical size={16} />
                </IconButton>
            </Tooltip>
            <Menu anchorEl={anchorElement} open={anchorElement !== null} onClose={closeMenu}>
                {entries.map((entry) => (
                    <MenuItem
                        key={entry.key}
                        onClick={() => selectEntry(entry)}
                        sx={entry.isDestructive ? { color: 'error.main' } : undefined}
                    >
                        <ListItemIcon sx={{ color: 'inherit' }}>{entry.icon}</ListItemIcon>
                        <ListItemText>{entry.label}</ListItemText>
                    </MenuItem>
                ))}
            </Menu>
        </>
    );
};

export default ActionBarMoreMenu;
