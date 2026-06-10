import { getMusicSectionMeta } from '@/features/music/components/navigation';
import { useMusicNavigation } from '@/features/music/components/useMusicNavigation';

export const useMusicDomainHeader = () => {
    const { currentSection } = useMusicNavigation();
    const section = getMusicSectionMeta(currentSection);

    return {
        titleKey: section.labelKey,
        descriptionKey: section.descriptionKey,
    };
};
