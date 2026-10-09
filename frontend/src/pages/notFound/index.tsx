import { Button } from '@mui/material';
import { useLocation, useNavigate } from 'react-router-dom';
import PageContainer from '@/components/layout/PageContainer/PageContainer';
import PageHeader from '@/components/layout/PageHeader/PageHeader';
import useI18n from '@/components/i18n/provider/i18nContext';
import { appRoutes } from '@/app/routes';

const NotFoundPage = () => {
    const { t } = useI18n();
    const { pathname } = useLocation();
    const navigate = useNavigate();

    const handleGoHome = () => navigate(appRoutes.home);
    const handleGoBack = () => navigate(-1);

    return (
        <PageContainer>
            <PageHeader
                title={t('NOT_FOUND_TITLE')}
                subtitle={t('NOT_FOUND_REQUESTED_PATH', { path: pathname })}
                actions={
                    <>
                        <Button variant="contained" onClick={handleGoHome}>
                            {t('GO_TO_HOME')}
                        </Button>
                        <Button variant="outlined" onClick={handleGoBack}>
                            {t('GO_BACK')}
                        </Button>
                    </>
                }
            />
        </PageContainer>
    );
};

export default NotFoundPage;
