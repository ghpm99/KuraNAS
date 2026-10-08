import type { ReactNode } from 'react';
import { Button, IconButton, Tooltip } from '@mui/material';

type ActionBarButtonProps = {
    label: string;
    icon: ReactNode;
    onClick: () => void;
    isIconOnly: boolean;
    isDestructive?: boolean;
    isPrimary?: boolean;
};

const ActionBarButton = ({
    label,
    icon,
    onClick,
    isIconOnly,
    isDestructive = false,
    isPrimary = false,
}: ActionBarButtonProps) => {
    const color = isDestructive ? 'error' : 'primary';

    if (isIconOnly) {
        return (
            <Tooltip title={label}>
                <IconButton size="small" color={color} aria-label={label} onClick={onClick}>
                    {icon}
                </IconButton>
            </Tooltip>
        );
    }

    return (
        <Button
            color={color}
            variant={isPrimary ? 'contained' : 'outlined'}
            size="small"
            startIcon={icon}
            onClick={onClick}
        >
            {label}
        </Button>
    );
};

export default ActionBarButton;
