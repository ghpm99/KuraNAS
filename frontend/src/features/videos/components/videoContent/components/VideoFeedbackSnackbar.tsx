import { Alert, Button, Snackbar } from '@mui/material';

export type VideoFeedbackAction = {
    label: string;
    onClick: () => void;
};

type VideoFeedbackSnackbarProps = {
    open: boolean;
    message: string;
    severity: 'success' | 'error';
    action?: VideoFeedbackAction;
    onClose: () => void;
};

export default function VideoFeedbackSnackbar({
    open,
    message,
    severity,
    action,
    onClose,
}: VideoFeedbackSnackbarProps) {
    const handleActionClick = () => {
        action?.onClick();
        onClose();
    };

    return (
        <Snackbar
            open={open}
            autoHideDuration={action ? 6000 : 2600}
            onClose={onClose}
            anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
        >
            <Alert
                severity={severity}
                variant="filled"
                onClose={onClose}
                action={
                    action ? (
                        <Button color="inherit" size="small" onClick={handleActionClick}>
                            {action.label}
                        </Button>
                    ) : undefined
                }
            >
                {message}
            </Alert>
        </Snackbar>
    );
}
