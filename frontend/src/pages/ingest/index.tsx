import ActiveIngestDownloads from '@/components/ingest/ActiveIngestDownloads';
import FinishedIngestDownloads from '@/components/ingest/FinishedIngestDownloads';
import IngestForm from '@/components/ingest/IngestForm';
import PageContainer from '@/components/layout/PageContainer';
import PageHeader from '@/components/layout/PageHeader';
import useI18n from '@/components/i18n/provider/i18nContext';

const IngestPage = () => {
    const { t } = useI18n();

    return (
        <PageContainer>
            <PageHeader title={t('INGEST_PAGE_TITLE')} subtitle={t('INGEST_PAGE_DESCRIPTION')} />
            <IngestForm />
            <ActiveIngestDownloads />
            <FinishedIngestDownloads />
        </PageContainer>
    );
};

export default IngestPage;
