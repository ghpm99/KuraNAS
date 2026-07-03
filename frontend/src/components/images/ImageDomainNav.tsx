import useI18n from '@/components/i18n/provider/i18nContext';
import DomainNavTabs from '@/components/layout/DomainNavTabs';
import { useImageNavigation } from '@/components/images/useImageNavigation';

const ImageDomainNav = () => {
    const { t } = useI18n();
    const { items } = useImageNavigation();

    return (
        <DomainNavTabs
            ariaLabel={t('IMAGES_NAVIGATION_LABEL')}
            items={items.map((item) => ({
                key: item.key,
                href: item.href,
                label: t(item.labelKey),
                isActive: item.isActive,
            }))}
        />
    );
};

export default ImageDomainNav;
