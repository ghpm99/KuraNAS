import { Alert, AlertTitle, Button } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';

type ErrorStateProps = {
    title: string;
    backendMessage?: string;
    onRetry?: () => void;
};

const ErrorState = ({ title, backendMessage, onRetry }: ErrorStateProps) => {
    const { t } = useI18n();

    return (
        <Alert
            severity="error"
            role="alert"
            action={
                onRetry ? (
                    <Button color="inherit" size="small" onClick={onRetry}>
                        {t('TRY_AGAIN')}
                    </Button>
                ) : undefined
            }
        >
            <AlertTitle>{title}</AlertTitle>
            {backendMessage}
        </Alert>
    );
};

export default ErrorState;
