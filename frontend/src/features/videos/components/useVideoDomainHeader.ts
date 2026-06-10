import { getVideoSectionMeta } from '@/features/videos/components/navigation';
import { useVideoNavigation } from '@/features/videos/components/useVideoNavigation';

export const useVideoDomainHeader = () => {
    const { currentSection } = useVideoNavigation();
    const section = getVideoSectionMeta(currentSection);

    return {
        titleKey: section.labelKey,
        descriptionKey: section.descriptionKey,
    };
};
