import useI18n from '@/components/i18n/provider/i18nContext';
import DomainNavTabs from '@/components/layout/DomainNavTabs';
import { useVideoNavigation } from '@/features/videos/components/useVideoNavigation';

const VideoDomainNav = () => {
    const { t } = useI18n();
    const { items } = useVideoNavigation();

    return (
        <DomainNavTabs
            ariaLabel={t('VIDEO_NAVIGATION_LABEL')}
            items={items.map((item) => ({
                key: item.key,
                href: item.href,
                label: t(item.labelKey),
                isActive: item.isActive,
            }))}
        />
    );
};

export default VideoDomainNav;
