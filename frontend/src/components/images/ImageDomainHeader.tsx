import useI18n from '@/components/i18n/provider/i18nContext';
import PageHeader from '@/components/layout/PageHeader';
import { useImageDomainHeader } from '@/components/images/useImageDomainHeader';

const ImageDomainHeader = () => {
    const { t } = useI18n();
    const { titleKey, descriptionKey } = useImageDomainHeader();

    return <PageHeader title={t(titleKey)} subtitle={t(descriptionKey)} />;
};

export default ImageDomainHeader;
