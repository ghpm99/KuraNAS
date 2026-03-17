import { getVideoSectionMeta } from '@/components/videos/navigation';
import { useVideoNavigation } from '@/components/videos/useVideoNavigation';

export const useVideoDomainHeader = () => {
    const { currentSection } = useVideoNavigation();
    const section = getVideoSectionMeta(currentSection);

    return {
        titleKey: section.labelKey,
        descriptionKey: section.descriptionKey,
    };
};
