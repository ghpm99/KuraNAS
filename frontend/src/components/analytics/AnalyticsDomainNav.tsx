import useI18n from '@/components/i18n/provider/i18nContext';
import DomainNavTabs from '@/components/layout/DomainNavTabs';
import { useAnalyticsNavigation } from '@/components/analytics/useAnalyticsNavigation';

const AnalyticsDomainNav = () => {
    const { t } = useI18n();
    const { items } = useAnalyticsNavigation();

    return (
        <DomainNavTabs
            ariaLabel={t('ANALYTICS_NAVIGATION_LABEL')}
            items={items.map((item) => ({
                key: item.key,
                href: item.href,
                label: t(item.labelKey),
                isActive: item.isActive,
            }))}
        />
    );
};

export default AnalyticsDomainNav;
