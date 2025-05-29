import { useQuery } from '@tanstack/react-query';
import { AboutContext, AboutContextType } from './AboutContext';
import { apiBase } from '@/service';

export function AboutProvider({ children }: { children: React.ReactNode }) {
	const { data } = useQuery({
		queryKey: ['about'],
		queryFn: async () => {
			const response = await apiBase.get<AboutContextType>('configuration/about');
			if (response.status !== 200) {
				throw new Error('Network response was not ok');
			}
			return response.data;
		},
		refetchOnWindowFocus: false,
	});

	return <AboutContext.Provider value={data}>{children}</AboutContext.Provider>;
}
