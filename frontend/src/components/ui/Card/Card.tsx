import { Card as MuiCard, CardContent, CardHeader } from '@mui/material';
import type { ReactNode } from 'react';

interface CardProps {
    title: string;
    children: ReactNode;
    className?: string;
}

export default function Card({ title, children }: CardProps) {
    return (
        <MuiCard>
            <CardHeader title={title} titleTypographyProps={{ variant: 'h6' }} />
            <CardContent>{children}</CardContent>
        </MuiCard>
    );
}
