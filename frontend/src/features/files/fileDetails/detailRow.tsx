import { Box, ListItem, Typography } from '@mui/material';
import type { ReactNode } from 'react';

type DetailRowProps = {
    label: string;
    value: ReactNode;
};

const DetailRow = ({ label, value }: DetailRowProps) => (
    <ListItem disablePadding sx={{ py: 0.5 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', gap: 1, width: '100%' }}>
            <Typography variant="caption" color="text.secondary">
                {label}
            </Typography>
            <Typography
                variant="caption"
                sx={{ maxWidth: '60%', textAlign: 'right', wordBreak: 'break-all' }}
            >
                {value}
            </Typography>
        </Box>
    </ListItem>
);

export default DetailRow;
