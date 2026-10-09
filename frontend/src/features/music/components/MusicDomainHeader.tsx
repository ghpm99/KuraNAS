import useI18n from '@/components/i18n/provider/i18nContext';
import PageHeader from '@/components/layout/PageHeader';
import MusicSearchField from '@/features/music/components/MusicSearchField';
import { useMusicDomainHeader } from '@/features/music/components/useMusicDomainHeader';

const MusicDomainHeader = () => {
    const { t } = useI18n();
    const { titleKey, descriptionKey } = useMusicDomainHeader();

    return (
        <PageHeader
            title={t(titleKey)}
            subtitle={t(descriptionKey)}
            actions={<MusicSearchField />}
        />
    );
};

export default MusicDomainHeader;
