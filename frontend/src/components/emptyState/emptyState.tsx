import type { ReactNode } from 'react';
import { Box, Typography } from '@mui/material';

type EmptyStateProps = {
    title: string;
    description?: string;
    icon?: ReactNode;
    actions?: ReactNode;
};

const EmptyState = ({ title, description, icon, actions }: EmptyStateProps) => (
    <Box
        role="status"
        sx={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 1.5,
            minHeight: 180,
            px: 2,
            py: 4,
            textAlign: 'center',
            border: '1px dashed var(--app-color-border-subtle)',
            borderRadius: 'var(--app-radius-lg)',
            color: 'var(--app-color-text-secondary)',
        }}
    >
        {icon}
        <Typography variant="subtitle1" fontWeight={600} color="text.primary">
            {title}
        </Typography>
        {description ? <Typography variant="body2">{description}</Typography> : null}
        {actions ? (
            <Box sx={{ display: 'flex', flexWrap: 'wrap', justifyContent: 'center', gap: 1 }}>
                {actions}
            </Box>
        ) : null}
    </Box>
);

export default EmptyState;
