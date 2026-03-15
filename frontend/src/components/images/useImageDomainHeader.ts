import { getImageSectionMeta } from '@/components/images/navigation';
import { useImageNavigation } from '@/components/images/useImageNavigation';

export const useImageDomainHeader = () => {
	const { currentSection } = useImageNavigation();
	const section = getImageSectionMeta(currentSection);

	return {
		titleKey: section.labelKey,
		descriptionKey: section.descriptionKey,
	};
};
