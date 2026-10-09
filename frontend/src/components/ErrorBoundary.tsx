import { Component, type ErrorInfo, type ReactNode } from 'react';
import { Box, Typography, Button } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import ErrorTechnicalDetails from '@/components/ErrorTechnicalDetails';

export interface ErrorFallbackProps {
    error: Error | null;
    onReset: () => void;
}

interface Props {
    children: ReactNode;
    resetKey?: string;
    renderFallback?: (fallbackProps: ErrorFallbackProps) => ReactNode;
}

const GlobalErrorFallback = ({ error, onReset }: ErrorFallbackProps) => {
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
                {t('ROUTE_ERROR_DESCRIPTION')}
            </Typography>
            <Button variant="contained" onClick={onReset}>
                {t('TRY_AGAIN')}
            </Button>
            <ErrorTechnicalDetails error={error} />
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

    componentDidUpdate(previousProps: Props) {
        const hasResetKeyChanged = previousProps.resetKey !== this.props.resetKey;
        if (this.state.hasError && hasResetKeyChanged) {
            this.handleReset();
        }
    }

    handleReset = () => {
        this.setState({ hasError: false, error: null });
    };

    render() {
        if (!this.state.hasError) return this.props.children;

        const fallbackProps = { error: this.state.error, onReset: this.handleReset };
        if (this.props.renderFallback) return this.props.renderFallback(fallbackProps);
        return <GlobalErrorFallback {...fallbackProps} />;
    }
}

export default ErrorBoundary;
