import { getAnalyticsSectionMeta } from '@/components/analytics/navigation';
import { useAnalyticsNavigation } from '@/components/analytics/useAnalyticsNavigation';

export const useAnalyticsDomainHeader = () => {
	const { currentSection } = useAnalyticsNavigation();
	const section = getAnalyticsSectionMeta(currentSection);

	return {
		titleKey: section.labelKey,
		descriptionKey: section.descriptionKey,
	};
};
