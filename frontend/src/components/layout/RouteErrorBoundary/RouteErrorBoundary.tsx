import type { ReactNode } from 'react';
import { Button } from '@mui/material';
import { useLocation, useNavigate } from 'react-router-dom';
import ErrorBoundary, { type ErrorFallbackProps } from '@/components/ErrorBoundary';
import ErrorTechnicalDetails from '@/components/ErrorTechnicalDetails';
import PageContainer from '@/components/layout/PageContainer/PageContainer';
import PageHeader from '@/components/layout/PageHeader/PageHeader';
import useI18n from '@/components/i18n/provider/i18nContext';
import { appRoutes } from '@/app/routes';

const RouteErrorFallback = ({ error, onReset }: ErrorFallbackProps) => {
    const { t } = useI18n();
    const navigate = useNavigate();
    const { pathname } = useLocation();

    const handleGoHome = () => {
        if (pathname === appRoutes.home) {
            onReset();
            return;
        }
        navigate(appRoutes.home);
    };

    return (
        <PageContainer>
            <PageHeader
                title={t('SOMETHING_WENT_WRONG')}
                subtitle={t('ROUTE_ERROR_DESCRIPTION')}
                actions={
                    <>
                        <Button variant="contained" onClick={onReset}>
                            {t('TRY_AGAIN')}
                        </Button>
                        <Button variant="outlined" onClick={handleGoHome}>
                            {t('GO_TO_HOME')}
                        </Button>
                    </>
                }
            />
            <ErrorTechnicalDetails error={error} />
        </PageContainer>
    );
};

const renderRouteErrorFallback = (fallbackProps: ErrorFallbackProps) => (
    <RouteErrorFallback {...fallbackProps} />
);

const RouteErrorBoundary = ({ children }: { children: ReactNode }) => {
    const { pathname } = useLocation();

    return (
        <ErrorBoundary resetKey={pathname} renderFallback={renderRouteErrorFallback}>
            {children}
        </ErrorBoundary>
    );
};

export default RouteErrorBoundary;
