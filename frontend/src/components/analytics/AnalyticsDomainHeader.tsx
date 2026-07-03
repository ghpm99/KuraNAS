import useI18n from '@/components/i18n/provider/i18nContext';
import PageHeader from '@/components/layout/PageHeader';
import { useAnalyticsDomainHeader } from '@/components/analytics/useAnalyticsDomainHeader';

const AnalyticsDomainHeader = () => {
    const { t } = useI18n();
    const { titleKey, descriptionKey } = useAnalyticsDomainHeader();

    return <PageHeader title={t(titleKey)} subtitle={t(descriptionKey)} />;
};

export default AnalyticsDomainHeader;
