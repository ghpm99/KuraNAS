import { Button as MuiButton } from '@mui/material';
import type { LucideIcon } from 'lucide-react';
import type { ButtonHTMLAttributes, ReactNode } from 'react';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
    children: ReactNode;
    variant?: 'primary' | 'secondary';
    icon?: LucideIcon;
}

export default function Button({
    children,
    variant = 'primary',
    icon: Icon,
    type,
    disabled,
    onClick,
    className,
}: ButtonProps) {
    return (
        <MuiButton
            variant={variant === 'primary' ? 'contained' : 'outlined'}
            startIcon={Icon ? <Icon size={16} /> : undefined}
            size="small"
            type={type as 'button' | 'submit' | 'reset'}
            disabled={disabled}
            onClick={onClick}
            className={className}
        >
            {children}
        </MuiButton>
    );
}
