import { Alert } from '@mui/material';

interface MessageProps {
    text: string;
    type: 'success' | 'error' | 'info';
}

export default function Message({ text, type }: MessageProps) {
    return <Alert severity={type}>{text}</Alert>;
}
