import useI18n from '@/components/i18n/provider/i18nContext';
import DomainNavTabs from '@/components/layout/DomainNavTabs';
import { useMusicNavigation } from '@/features/music/components/useMusicNavigation';

const MusicDomainNav = () => {
    const { t } = useI18n();
    const { items } = useMusicNavigation();

    return (
        <DomainNavTabs
            ariaLabel={t('MUSIC_NAVIGATION_LABEL')}
            items={items.map((item) => ({
                key: item.key,
                href: item.href,
                label: t(item.labelKey),
                isActive: item.isActive,
            }))}
        />
    );
};

export default MusicDomainNav;
