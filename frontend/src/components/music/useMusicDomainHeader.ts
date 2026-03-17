import { getMusicSectionMeta } from '@/components/music/navigation';
import { useMusicNavigation } from '@/components/music/useMusicNavigation';

export const useMusicDomainHeader = () => {
    const { currentSection } = useMusicNavigation();
    const section = getMusicSectionMeta(currentSection);

    return {
        titleKey: section.labelKey,
        descriptionKey: section.descriptionKey,
    };
};
