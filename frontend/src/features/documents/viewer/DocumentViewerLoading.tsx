import { Box, CircularProgress, Typography } from '@mui/material';

const DocumentViewerLoading = ({ label }: { label: string }) => (
    <Box role="status" sx={{ display: 'flex', alignItems: 'center', gap: 1, p: 2 }}>
        <CircularProgress size={16} />
        <Typography variant="body2">{label}</Typography>
    </Box>
);

export default DocumentViewerLoading;
