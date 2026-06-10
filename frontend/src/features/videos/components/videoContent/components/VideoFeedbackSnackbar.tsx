import { Alert, Snackbar } from '@mui/material';

type VideoFeedbackSnackbarProps = {
    open: boolean;
    message: string;
    severity: 'success' | 'error';
    onClose: () => void;
};

export default function VideoFeedbackSnackbar({
    open,
    message,
    severity,
    onClose,
}: VideoFeedbackSnackbarProps) {
    return (
        <Snackbar
            open={open}
            autoHideDuration={2600}
            onClose={onClose}
            anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
        >
            <Alert severity={severity} variant="filled" onClose={onClose}>
                {message}
            </Alert>
        </Snackbar>
    );
}
