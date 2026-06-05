import { Component, type ErrorInfo, type ReactNode } from 'react';
import { Box, Typography, Button } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';

interface Props {
    children: ReactNode;
}

const ErrorFallback = ({ message, onReset }: { message?: string; onReset: () => void }) => {
    const { t } = useI18n();
    return (
        <Box
            sx={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                minHeight: '100vh',
                gap: 2,
                p: 4,
            }}
        >
            <Typography variant="h5">{t('SOMETHING_WENT_WRONG')}</Typography>
            <Typography
                variant="body2"
                color="text.secondary"
                sx={{ maxWidth: 600, textAlign: 'center' }}
            >
                {message}
            </Typography>
            <Button variant="contained" onClick={onReset}>
                {t('TRY_AGAIN')}
            </Button>
        </Box>
    );
};

interface State {
    hasError: boolean;
    error: Error | null;
}

class ErrorBoundary extends Component<Props, State> {
    constructor(props: Props) {
        super(props);
        this.state = { hasError: false, error: null };
    }

    static getDerivedStateFromError(error: Error): State {
        return { hasError: true, error };
    }

    componentDidCatch(error: Error, errorInfo: ErrorInfo) {
        console.error('ErrorBoundary caught:', error, errorInfo);
    }

    handleReset = () => {
        this.setState({ hasError: false, error: null });
    };

    render() {
        if (this.state.hasError) {
            return (
                <ErrorFallback
                    message={this.state.error?.message}
                    onReset={this.handleReset}
                />
            );
        }

        return this.props.children;
    }
}

export default ErrorBoundary;
