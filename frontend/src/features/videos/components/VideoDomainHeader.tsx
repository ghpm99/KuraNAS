import useI18n from '@/components/i18n/provider/i18nContext';
import PageHeader from '@/components/layout/PageHeader';
import { useVideoDomainHeader } from '@/features/videos/components/useVideoDomainHeader';

const VideoDomainHeader = () => {
    const { t } = useI18n();
    const { titleKey, descriptionKey } = useVideoDomainHeader();

    return <PageHeader title={t(titleKey)} subtitle={t(descriptionKey)} />;
};

export default VideoDomainHeader;
