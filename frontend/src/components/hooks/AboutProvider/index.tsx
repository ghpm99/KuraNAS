import { useQuery } from '@tanstack/react-query';
import { AboutContext, AboutContextType } from './AboutContext';
import { apiBase } from '@/service';

const initialAboutContext: AboutContextType = {
	version: '',
	commit_hash: '',
	platform: '',
	enable_workers: false,
	gin_mode: '',
	lang: '',
	path: '',
	statup_time: new Date().toISOString(),
	gin_version: '',
};

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

	const value = data || initialAboutContext;

	return <AboutContext.Provider value={value}>{children}</AboutContext.Provider>;
}
